package cmd

import (
	"fmt"
	"log"

	"git-fs/internal/config"
	"git-fs/internal/crypto"
	fileutils "git-fs/internal/fileutil"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the repository and encryption",
	Long:  `Sets up the repository, generates a salt, and derives an encryption key.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		saltPath := cfg.RepoPath + "/.salt"
		salt, err := fileutils.ReadOrCreateSalt(saltPath)
		if err != nil {
			log.Fatalf("Failed to create/read salt: %v", err)
		}

		_, err = crypto.DeriveKey(cfg.Password, salt)
		if err != nil {
			log.Fatalf("Error deriving key: %v", err)
		}

		fmt.Println("Repository initialized with encryption key.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
