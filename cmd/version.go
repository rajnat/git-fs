// cmd/version.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
		fmt.Printf("git-fs version %s (commit %s)\n", version, commit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
