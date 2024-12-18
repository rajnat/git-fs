package main

import (
	"git-fs/cmd"
	"git-fs/internal/logging"
	"os"
)

func main() {
	debug := os.Getenv("GITFS_DEBUG") == "true"
	if err := logging.InitLogger(debug); err != nil {
		// If logger initialization fails, just use the standard log as a fallback.
		panic("failed to initialize logger: " + err.Error())
	}
	defer logging.Sync()

	cmd.Execute()
}
