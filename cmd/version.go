package cmd

import (
	"git-fs/internal/logging"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	version = "0.1.0"
	commit  = "abc123"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Displays the build version and commit of git-fs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Using zap's structured logging to report the version and commit.
		logging.Logger.Info("git-fs version information",
			zap.String("version", version),
			zap.String("commit", commit),
		)
		// Optionally print to stdout if you want the user to see it directly:
		// cmd.Println(fmt.Sprintf("git-fs version %s (commit %s)", version, commit))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
