// cmd/daemon.go
package cmd

import (
	"fmt"
	"git-fs/internal/config"
	"log"

	"github.com/spf13/cobra"
	// Add imports for watcher logic, fsnotify, etc.
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the watcher daemon",
	Long:  `Runs in the background, watching the directory for changes and encrypting & committing them.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		fmt.Printf("Starting daemon. Repo: %s, Watch: %s\n", cfg.RepoPath, cfg.WatchPath)
		// Implement watcher logic here.
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
