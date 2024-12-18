package cmd

import (
	"git-fs/internal/config"
	"git-fs/internal/logging"
	"git-fs/internal/status"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current status of the daemon and repository",
	Long:  `Displays whether the watcher is running, how many files are pending, and details of the last commit/push.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.Logger

		// We load config to know where the repo is:
		cfg, err := config.LoadConfig()
		if err != nil {
			logger.Error("Failed to load config", zap.Error(err))
			cmd.PrintErrln("Error: Could not load configuration.")
			return
		}

		statusPath := filepath.Join(cfg.RepoPath, ".status.json")
		st, err := status.LoadStatus(statusPath)
		if err != nil {
			logger.Error("Failed to load status", zap.String("path", statusPath), zap.Error(err))
			cmd.PrintErrln("Error: Could not load status. Is the daemon running?")
			return
		}

		// Print out status in a user-friendly format
		cmd.Println("git-fs Status:")
		cmd.Printf("  Watcher running: %v\n", st.WatcherRunning)
		cmd.Printf("  Files pending encryption: %d\n", st.FilesPending)

		if st.LastCommitHash == "" {
			cmd.Println("  No commits recorded yet.")
		} else {
			cmd.Printf("  Last commit: %s at %s\n", st.LastCommitHash, st.LastCommitTime.Format(time.RFC3339))
		}

		cmd.Printf("  Last push successful: %v\n", st.LastPushSuccessful)
		if !st.LastPushTime.IsZero() {
			cmd.Printf("  Last push time: %s\n", st.LastPushTime.Format(time.RFC3339))
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
