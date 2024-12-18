package cmd

import (
	"fmt"
	"log"

	"git-fs/internal/config"
	"git-fs/internal/daemon"

	"github.com/spf13/cobra"
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
		err = daemon.RunDaemon(cfg)
		if err != nil {
			log.Fatalf("Failed to run daemon: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
