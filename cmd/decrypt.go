// cmd/decrypt.go
package cmd

import (
	"log"
	"path/filepath"
	"strings"

	"git-fs/internal/config"
	"git-fs/internal/crypto"
	fileutils "git-fs/internal/fileutil"

	"github.com/spf13/cobra"
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt the encrypted files in the repository",
	Long:  `Takes the files from the .encrypted directory and decrypts them to their original form.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}

		saltPath := filepath.Join(cfg.RepoPath, ".salt")
		salt, err := fileutils.SafeReadFile(saltPath)
		if err != nil {
			log.Fatalf("Failed to read salt: %v", err)
		}

		key, err := crypto.DeriveKey(cfg.Password, salt)
		if err != nil {
			log.Fatalf("Error deriving key: %v", err)
		}

		encryptedRoot := filepath.Join(cfg.RepoPath, ".encrypted")
		files, err := fileutils.GetFiles(encryptedRoot)
		if err != nil {
			log.Fatalf("Error getting encrypted files: %v", err)
		}

		for _, encFile := range files {
			rel, _ := filepath.Rel(encryptedRoot, encFile)
			plaintextRel := strings.TrimSuffix(rel, ".enc")
			plaintextPath := filepath.Join(cfg.RepoPath, plaintextRel)

			if err := fileutils.EnsureDir(filepath.Dir(plaintextPath)); err != nil {
				log.Fatalf("Failed to ensure directory for %s: %v", plaintextPath, err)
			}
			if err := crypto.DecryptFile(key, encFile, plaintextPath); err != nil {
				log.Fatalf("Failed to decrypt file %s: %v", encFile, err)
			}
		}
		log.Println("Decryption complete.")
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)
}
