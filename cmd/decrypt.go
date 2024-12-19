package cmd

import (
	"git-fs/internal/config"
	"git-fs/internal/crypto"
	filemetadata "git-fs/internal/filemetadata"
	fileutils "git-fs/internal/fileutil"
	"git-fs/internal/logging"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt the encrypted files in the repository",
	Long:  `Takes the files from the .encrypted directory and decrypts them to their original form.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.Logger

		cfg, err := config.LoadConfig()
		if err != nil {
			logger.Error("Error loading config", zap.Error(err))
			cmd.PrintErrln("Error: Could not load configuration. Please ensure config.yaml or ENV variables are set.")
			return
		}

		saltPath := filepath.Join(cfg.RepoPath, ".salt")
		salt, err := fileutils.SafeReadFile(saltPath)
		if err != nil {
			logger.Error("Failed to read salt", zap.String("path", saltPath), zap.Error(err))
			cmd.PrintErrln("Error: Failed to read the salt file. Ensure the repository is initialized and the .salt file is present.")
			return
		}

		key, err := crypto.DeriveKey(cfg.Password, salt)
		if err != nil {
			logger.Error("Error deriving key", zap.Error(err))
			cmd.PrintErrln("Error: Unable to derive encryption key. Check your password and try again.")
			return
		}

		// Load metadata store
		metadataPath := filepath.Join(cfg.RepoPath, ".metadata.enc")
		metadataStore, err := filemetadata.LoadMetadataStore(metadataPath, key)
		if err != nil {
			logger.Error("Failed to load metadata store", zap.Error(err))
			cmd.PrintErrln("Error: Could not load metadata store. Ensure the file exists and the password is correct.")
			return
		}

		encryptedRoot := filepath.Join(cfg.RepoPath, ".encrypted")

		// Process each file in the metadata store
		for encryptedName, metadata := range metadataStore.Metadata {
			encryptedPath := filepath.Join(encryptedRoot, encryptedName)

			// Skip if encrypted file doesn't exist
			if !fileutils.FileExists(encryptedPath) {
				logger.Warn("Encrypted file not found",
					zap.String("encrypted_path", encryptedPath))
				continue
			}

			// Create the output directory structure
			outputPath := filepath.Join(cfg.WatchPath, metadata.OriginalPath)
			if err := fileutils.EnsureDir(filepath.Dir(outputPath)); err != nil {
				logger.Error("Failed to create directory",
					zap.String("path", filepath.Dir(outputPath)),
					zap.Error(err))
				continue
			}

			// Decrypt the file
			if err := crypto.DecryptFile(key, encryptedPath, outputPath); err != nil {
				logger.Error("Failed to decrypt file",
					zap.String("encrypted_file", encryptedPath),
					zap.String("output_path", outputPath),
					zap.Error(err))
				continue
			}

			logger.Info("File decrypted",
				zap.String("encrypted_file", encryptedPath),
				zap.String("decrypted_file", outputPath))
		}

		logger.Info("Decryption complete", zap.String("watch_path", cfg.WatchPath))
		cmd.Println("Decryption complete.")
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)
}
