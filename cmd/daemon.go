package cmd

import (
	"git-fs/internal/config"
	"git-fs/internal/daemon"
	"git-fs/internal/logging"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the watcher daemon",
	Long:  `Runs in the background, watching the directory for changes and encrypting & committing them.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.Logger

		cfg, err := config.LoadConfig()
		if err != nil {
			logger.Error("Failed to load config", zap.Error(err))
			cmd.Println("Error: Could not load configuration. Please ensure config.yaml or ENV variables are set.")
			return
		}

		logger.Info("Starting daemon", zap.String("repo", cfg.RepoPath), zap.String("watch", cfg.WatchPath))

		if err := daemon.RunDaemon(cfg); err != nil {
			logger.Error("Failed to run daemon", zap.Error(err))
			cmd.Println("Error: Daemon failed to start. Check logs for details.")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
