package daemon

import (
	"crypto/sha256"
	"encoding/base64"

	"errors"
	"os"
	"path/filepath"

	"time"

	"git-fs/internal/config"
	"git-fs/internal/crypto"
	filemetadata "git-fs/internal/filemetadata"
	fileutils "git-fs/internal/fileutil"
	"git-fs/internal/gitutils"
	"git-fs/internal/logging"
	"git-fs/internal/status"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
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

	// Load or create metadata store
	metadataPath := filepath.Join(cfg.RepoPath, ".metadata.enc")
	metadataStore, err := filemetadata.LoadMetadataStore(metadataPath, key)
	if err != nil {
		logger.Error("Failed to load metadata store", zap.Error(err))
		return errors.New("could not load metadata store")
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

	st := &status.Status{
		WatcherRunning: true,
	}
	statusPath := filepath.Join(cfg.RepoPath, ".status.json")
	status.SaveStatus(statusPath, st)

	cs := &filemetadata.ChangeSet{Files: make(map[string]struct{})}

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
					logger.Info("Watcher events channel closed, stopping daemon.")
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					cs.Mu.Lock()
					cs.Files[event.Name] = struct{}{}
					cs.Mu.Unlock()

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

	// debounce goroutine to include metadata handling
	go func() {
		for {
			<-debounce.C
			cs.Mu.Lock()
			changedFiles := make([]string, 0, len(cs.Files))
			for f := range cs.Files {
				changedFiles = append(changedFiles, f)
			}
			cs.Files = make(map[string]struct{})
			cs.Mu.Unlock()

			if len(changedFiles) > 0 {
				logger.Info("Processing changes", zap.Int("file_count", len(changedFiles)))
				if err := handleChanges(cfg, key, changedFiles, st, statusPath, metadataStore, metadataPath); err != nil {
					logger.Error("Failed to handle changes", zap.Error(err))
				}
			}
		}
	}()

	<-done
	return nil
}

func handleChanges(cfg *config.Config, key []byte, changedFiles []string, st *status.Status,
	statusPath string, metadataStore *filemetadata.MetadataStore, metadataPath string) error {
	logger := logging.Logger
	encryptedRoot := filepath.Join(cfg.RepoPath, ".encrypted")

	st.FilesPending = len(changedFiles)
	if err := status.SaveStatus(statusPath, st); err != nil {
		logger.Warn("Failed to save status before encryption", zap.Error(err))
	}

	for _, f := range changedFiles {
		fileInfo, err := fileutils.SafeStat(f)
		if err != nil {
			// Handle deleted files
			relPath, _ := filepath.Rel(cfg.WatchPath, f)
			metadataStore.Mu.Lock()
			for encName, metadata := range metadataStore.Metadata {
				if metadata.OriginalPath == relPath {
					encPath := filepath.Join(encryptedRoot, encName)
					if fileutils.FileExists(encPath) {
						if removeErr := os.Remove(encPath); removeErr != nil {
							logger.Warn("Failed to remove encrypted file",
								zap.String("path", encPath), zap.Error(removeErr))
						}
					}
					delete(metadataStore.Metadata, encName)
					break
				}
			}
			metadataStore.Mu.Unlock()
			st.FilesPending--
			status.SaveStatus(statusPath, st)
			continue
		}

		if !fileInfo.IsDir() {
			// Read file content
			content, err := os.ReadFile(f)
			if err != nil {
				logger.Error("Failed to read file", zap.String("file", f), zap.Error(err))
				st.FilesPending--
				status.SaveStatus(statusPath, st)
				continue
			}

			// Calculate original hash
			originalHash := calculateHash(content)

			// Generate encrypted filename
			relPath, _ := filepath.Rel(cfg.WatchPath, f)
			encryptedName, fileNonce, nameNonce, err := crypto.EncryptFileName(key, relPath)
			if err != nil {
				logger.Error("Failed to encrypt filename", zap.String("file", f), zap.Error(err))
				st.FilesPending--
				status.SaveStatus(statusPath, st)
				continue
			}

			encPath := filepath.Join(encryptedRoot, encryptedName)
			if err := fileutils.EnsureDir(filepath.Dir(encPath)); err != nil {
				logger.Error("Failed to ensure directory",
					zap.String("path", filepath.Dir(encPath)),
					zap.Error(err))
				st.FilesPending--
				status.SaveStatus(statusPath, st)
				continue
			}

			// Encrypt file content
			encryptedContent, err := crypto.EncryptFile(key, content, fileNonce)
			if err != nil {
				logger.Error("Failed to encrypt file",
					zap.String("file", f),
					zap.String("encrypted_path", encPath),
					zap.Error(err))
				st.FilesPending--
				status.SaveStatus(statusPath, st)
				continue
			}

			// Calculate encrypted hash
			encryptedHash := calculateHash(encryptedContent)

			// Save encrypted content
			if err := os.WriteFile(encPath, encryptedContent, 0600); err != nil {
				logger.Error("Failed to write encrypted file",
					zap.String("path", encPath),
					zap.Error(err))
				st.FilesPending--
				status.SaveStatus(statusPath, st)
				continue
			}

			// Update metadata
			metadata := filemetadata.FileMetadata{
				EncryptedName:   encryptedName,
				OriginalPath:    relPath,
				OriginalHash:    originalHash,
				EncryptedHash:   encryptedHash,
				LastModified:    time.Now(),
				FileSize:        fileInfo.Size(),
				EncryptionNonce: nameNonce,
				FileNonce:       fileNonce,
			}

			metadataStore.Mu.Lock()
			metadataStore.Metadata[encryptedName] = metadata
			metadataStore.Mu.Unlock()

			logger.Info("File encrypted",
				zap.String("file", f),
				zap.String("encrypted", encPath))
			st.FilesPending--
			status.SaveStatus(statusPath, st)
		} else {
			st.FilesPending--
			status.SaveStatus(statusPath, st)
		}
	}

	// Save metadata before git operations
	if err := metadataStore.SaveToFile(metadataPath, key); err != nil {
		logger.Error("Failed to save metadata", zap.Error(err))
		return err
	}

	// Add both encrypted files and metadata to git
	if err := gitutils.AddAndCommit(cfg.RepoPath, "Automated encrypted backup"); err == nil {
		if hash, cerr := gitutils.GetLastCommitHash(cfg.RepoPath); cerr == nil {
			st.LastCommitHash = hash
			st.LastCommitTime = time.Now()
		}
		st.FilesPending = 0
		status.SaveStatus(statusPath, st)
	} else {
		logger.Error("Git commit failed", zap.Error(err))
		return errors.New("git commit failed; ensure you have a valid repo and permissions")
	}

	if cfg.RemoteURL != "" {
		if err := gitutils.Push(cfg.RepoPath, "origin", "main"); err != nil {
			logger.Error("Failed to push to remote",
				zap.String("remote_url", cfg.RemoteURL),
				zap.Error(err))
			return errors.New("failed to push to remote repository; check your network or remote configuration")
		} else {
			st.LastPushSuccessful = true
			st.LastPushTime = time.Now()
			status.SaveStatus(statusPath, st)
			logger.Info("Changes pushed to remote",
				zap.String("remote_url", cfg.RemoteURL))
		}
	}

	return nil
}
