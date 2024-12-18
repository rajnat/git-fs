package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"git-fs/internal/config"
	"git-fs/internal/crypto"
	fileutils "git-fs/internal/fileutil"
	"git-fs/internal/gitutils"

	"github.com/fsnotify/fsnotify"
)

// We'll keep some internal state
type changeSet struct {
	mu    sync.Mutex
	files map[string]struct{}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Derive the encryption key using a stable salt
	// In a real scenario, you’d have stored the salt in the repo previously:
	saltPath := filepath.Join(cfg.RepoPath, ".salt")
	salt, err := fileutils.ReadOrCreateSalt(saltPath)
	if err != nil {
		log.Fatalf("Failed to get salt: %v", err)
	}

	key, err := crypto.DeriveKey(cfg.Password, salt)
	if err != nil {
		log.Fatalf("Error deriving key: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creating watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(cfg.WatchPath)
	if err != nil {
		log.Fatalf("Error adding watch path: %v", err)
	}

	cs := &changeSet{files: make(map[string]struct{})}

	// Debounce mechanism:
	debounce := time.NewTimer(0)
	if !debounce.Stop() {
		<-debounce.C
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Track changes
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					cs.mu.Lock()
					cs.files[event.Name] = struct{}{}
					cs.mu.Unlock()
					// Reset the debounce timer
					if !debounce.Stop() {
						select {
						case <-debounce.C:
						default:
						}
					}
					debounce.Reset(2 * time.Second)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	go func() {
		for {
			<-debounce.C
			// Time to handle all changes accumulated
			cs.mu.Lock()
			changedFiles := make([]string, 0, len(cs.files))
			for f := range cs.files {
				changedFiles = append(changedFiles, f)
			}
			cs.files = make(map[string]struct{})
			cs.mu.Unlock()

			if len(changedFiles) > 0 {
				// Encrypt and commit changes
				handleChanges(cfg, key, changedFiles)
			}
		}
	}()

	<-done // Keep the daemon running indefinitely
}

func handleChanges(cfg *config.Config, key []byte, changedFiles []string) {
	// Not all changed files might need encryption.
	// We only encrypt files that are inside WATCH_PATH and not directories.
	// Remove or handle deleted files as needed.

	encryptedRoot := filepath.Join(cfg.RepoPath, ".encrypted")

	for _, f := range changedFiles {
		fileInfo, err := fileutils.SafeStat(f)
		if err != nil {
			// File might have been removed.
			// Consider how to handle removals:
			// If file is removed, we could also remove the encrypted version.
			encPath := encryptedFilePath(cfg, f, encryptedRoot)
			if fileutils.FileExists(encPath) {
				os.Remove(encPath)
			}
			continue
		}
		if !fileInfo.IsDir() {
			// Encrypt this file
			encPath := encryptedFilePath(cfg, f, encryptedRoot)
			if err := fileutils.EnsureDir(filepath.Dir(encPath)); err != nil {
				log.Printf("Failed to ensure directory for %s: %v", encPath, err)
				continue
			}
			if err := crypto.EncryptFile(key, f, encPath); err != nil {
				log.Printf("Failed to encrypt file %s: %v", f, err)
				continue
			}
		}
	}

	// Commit changes to git
	if err := gitutils.AddAndCommit(cfg.RepoPath, "Automated encrypted backup"); err != nil {
		log.Printf("Git commit failed: %v", err)
	} else {
		// Optionally push to remote if REMOTE_REPO is configured
		if cfg.RemoteURL != "" {
			if err := gitutils.Push(cfg.RepoPath, "origin", "main"); err != nil {
				log.Printf("Failed to push: %v", err)
			}
		}
	}
}

func encryptedFilePath(cfg *config.Config, filePath, encryptedRoot string) string {
	rel, err := filepath.Rel(cfg.WatchPath, filePath)
	if err != nil {
		return filepath.Join(encryptedRoot, filepath.Base(filePath)+".enc")
	}
	return filepath.Join(encryptedRoot, rel+".enc")
}
