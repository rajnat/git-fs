package daemon

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"git-fs/internal/config"
	"git-fs/internal/crypto"
	fileutils "git-fs/internal/fileutil"
	"git-fs/internal/gitutils"
	"git-fs/internal/logging"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type changeSet struct {
	mu    sync.Mutex
	files map[string]struct{}
}

func RunDaemon(cfg *config.Config) error {
	logger := logging.Logger

	saltPath := filepath.Join(cfg.RepoPath, ".salt")
	salt, err := fileutils.ReadOrCreateSalt(saltPath)
	if err != nil {
		logger.Error("Failed to get or create salt", zap.String("path", saltPath), zap.Error(err))
		return errors.New("could not initialize encryption salt; please check permissions or run `git-fs init` first")
	}

	key, err := crypto.DeriveKey(cfg.Password, salt)
	if err != nil {
		logger.Error("Failed to derive key", zap.Error(err))
		return errors.New("invalid password or salt; cannot derive encryption key")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("Failed to create watcher", zap.Error(err))
		return errors.New("could not create file watcher")
	}
	defer watcher.Close()

	if err = watcher.Add(cfg.WatchPath); err != nil {
		logger.Error("Failed to add watch path", zap.String("watchPath", cfg.WatchPath), zap.Error(err))
		return errors.New("could not watch the specified directory; please check if it exists and is accessible")
	}

	cs := &changeSet{files: make(map[string]struct{})}

	debounce := time.NewTimer(0)
	if !debounce.Stop() {
		<-debounce.C
	}

	done := make(chan bool)

	// Watcher goroutine
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					logger.Info("Watcher events channel closed, stopping daemon.")
					return
				}
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

			case werr, ok := <-watcher.Errors:
				if !ok {
					logger.Warn("Watcher errors channel closed")
					return
				}
				logger.Error("Watcher error occurred", zap.Error(werr))
			}
		}
	}()

	// Debounce goroutine
	go func() {
		for {
			<-debounce.C
			cs.mu.Lock()
			changedFiles := make([]string, 0, len(cs.files))
			for f := range cs.files {
				changedFiles = append(changedFiles, f)
			}
			cs.files = make(map[string]struct{})
			cs.mu.Unlock()

			if len(changedFiles) > 0 {
				logger.Info("Processing changes", zap.Int("file_count", len(changedFiles)))
				if err := handleChanges(cfg, key, changedFiles); err != nil {
					logger.Error("Failed to handle changes", zap.Error(err))
				}
			}
		}
	}()

	<-done // This will block forever unless you have logic to close done
	return nil
}

func handleChanges(cfg *config.Config, key []byte, changedFiles []string) error {
	logger := logging.Logger
	encryptedRoot := filepath.Join(cfg.RepoPath, ".encrypted")

	for _, f := range changedFiles {
		fileInfo, err := fileutils.SafeStat(f)
		if err != nil {
			// File might have been removed
			encPath := encryptedFilePath(cfg, f, encryptedRoot)
			if fileutils.FileExists(encPath) {
				if removeErr := os.Remove(encPath); removeErr != nil {
					logger.Warn("Failed to remove encrypted file for deleted plaintext",
						zap.String("path", encPath), zap.Error(removeErr))
				} else {
					logger.Info("Removed encrypted file for deleted plaintext",
						zap.String("path", encPath))
				}
			}
			continue
		}

		if !fileInfo.IsDir() {
			// Encrypt this file
			encPath := encryptedFilePath(cfg, f, encryptedRoot)
			if err := fileutils.EnsureDir(filepath.Dir(encPath)); err != nil {
				logger.Error("Failed to ensure directory",
					zap.String("path", filepath.Dir(encPath)),
					zap.Error(err))
				continue
			}
			if err := crypto.EncryptFile(key, f, encPath); err != nil {
				logger.Error("Failed to encrypt file",
					zap.String("file", f),
					zap.String("encrypted_path", encPath),
					zap.Error(err))
				continue
			}
			logger.Info("File encrypted",
				zap.String("file", f),
				zap.String("encrypted", encPath))
		}
	}

	if err := gitutils.AddAndCommit(cfg.RepoPath, "Automated encrypted backup"); err != nil {
		logger.Error("Git commit failed", zap.Error(err))
		return errors.New("git commit failed; ensure you have a valid repo and permissions")
	}

	// Optionally push to remote
	if cfg.RemoteURL != "" {
		if err := gitutils.Push(cfg.RepoPath, "origin", "main"); err != nil {
			logger.Error("Failed to push to remote", zap.String("remote_url", cfg.RemoteURL), zap.Error(err))
			return errors.New("failed to push to remote repository; check your network or remote configuration")
		}
		logger.Info("Changes pushed to remote",
			zap.String("remote_url", cfg.RemoteURL))
	}

	return nil
}

func encryptedFilePath(cfg *config.Config, filePath, encryptedRoot string) string {
	rel, err := filepath.Rel(cfg.WatchPath, filePath)
	if err != nil {
		// If we cannot get relative path, fallback to a safe name
		return filepath.Join(encryptedRoot, filepath.Base(filePath)+".enc")
	}
	return filepath.Join(encryptedRoot, rel+".enc")
}
