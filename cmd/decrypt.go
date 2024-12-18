package cmd

import (
	"path/filepath"
	"strings"

	"git-fs/internal/config"
	"git-fs/internal/crypto"
	fileutils "git-fs/internal/fileutil"
	"git-fs/internal/logging"

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

		encryptedRoot := filepath.Join(cfg.RepoPath, ".encrypted")
		files, err := fileutils.GetFiles(encryptedRoot)
		if err != nil {
			logger.Error("Error getting encrypted files", zap.String("encrypted_root", encryptedRoot), zap.Error(err))
			cmd.PrintErrln("Error: Could not find encrypted files. Ensure your repository is correctly initialized and contains `.encrypted` directory.")
			return
		}

		for _, encFile := range files {
			rel, _ := filepath.Rel(encryptedRoot, encFile)
			plaintextRel := strings.TrimSuffix(rel, ".enc")
			plaintextPath := filepath.Join(cfg.RepoPath, plaintextRel)

			if err := fileutils.EnsureDir(filepath.Dir(plaintextPath)); err != nil {
				logger.Error("Failed to ensure directory",
					zap.String("directory", filepath.Dir(plaintextPath)),
					zap.Error(err))
				cmd.PrintErrln("Error: Failed to create necessary directories for decrypted files. Check your file system permissions.")
				return
			}

			if err := crypto.DecryptFile(key, encFile, plaintextPath); err != nil {
				logger.Error("Failed to decrypt file",
					zap.String("encrypted_file", encFile),
					zap.String("plaintext_path", plaintextPath),
					zap.Error(err))
				cmd.PrintErrln("Error: Failed to decrypt one or more files. Ensure the correct key and file integrity.")
				return
			}

			logger.Info("File decrypted",
				zap.String("encrypted_file", encFile),
				zap.String("decrypted_file", plaintextPath))
		}

		logger.Info("Decryption complete", zap.String("repo_path", cfg.RepoPath))
		cmd.Println("Decryption complete.")
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)
}
