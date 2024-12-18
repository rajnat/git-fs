package cmd

import (
	"git-fs/internal/config"
	"git-fs/internal/crypto"
	fileutils "git-fs/internal/fileutil"
	"git-fs/internal/logging"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the repository and encryption",
	Long:  `Sets up the repository, generates a salt, and derives an encryption key.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.Logger

		cfg, err := config.LoadConfig()
		if err != nil {
			logger.Error("Failed to load config", zap.Error(err))
			cmd.PrintErrln("Error: Could not load configuration. Please ensure config.yaml or ENV variables are set.")
			return
		}

		saltPath := cfg.RepoPath + "/.salt"
		salt, err := fileutils.ReadOrCreateSalt(saltPath)
		if err != nil {
			logger.Error("Failed to create/read salt", zap.String("path", saltPath), zap.Error(err))
			cmd.PrintErrln("Error: Failed to initialize repository salt. Ensure the repo path is correct and writable.")
			return
		}

		if _, err := crypto.DeriveKey(cfg.Password, salt); err != nil {
			logger.Error("Error deriving key", zap.Error(err))
			cmd.PrintErrln("Error: Unable to derive encryption key. Check your password and try again.")
			return
		}

		logger.Info("Repository initialized with encryption key", zap.String("repo_path", cfg.RepoPath))
		cmd.Println("Repository initialized with encryption key.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
