package main

import (
	"flag"
	"fmt"
	"git-fs/internal/config"
	"git-fs/internal/crypto"
	fileutils "git-fs/internal/fileutil"
	"git-fs/internal/gitutils"
	"log"
	"path/filepath"
)

func main() {
	folderPath := flag.String("path", ".", "Path to the folder to encrypt")
	message := flag.String("message", "Encrypted files update", "Git commit message")
	flag.Parse()

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading the config: %v", err)
	}
	var password = config.Password
	salt, err := crypto.GenerateSalt()
	if err != nil {
		log.Fatalf("Error generating salt: %v", err)
	}

	key, err := crypto.DeriveKey(password, salt)
	if err != nil {
		log.Fatalf("Error deriving key: %v", err)
	}

	files, err := fileutils.GetFiles(*folderPath)
	if err != nil {
		log.Fatalf("Error getting files: %v", err)
	}

	// Example scheme: Encrypted files go into a parallel structure under .encrypted/
	encryptedRoot := filepath.Join(*folderPath, ".encrypted")
	if err := fileutils.EnsureDir(encryptedRoot); err != nil {
		log.Fatalf("Error ensuring encrypted directory: %v", err)
	}

	for _, f := range files {
		rel, _ := filepath.Rel(*folderPath, f)
		encryptedPath := filepath.Join(encryptedRoot, rel+".enc")
		if err := fileutils.EnsureDir(filepath.Dir(encryptedPath)); err != nil {
			log.Fatalf("Failed to ensure directory for %s: %v", encryptedPath, err)
		}
		if err := crypto.EncryptFile(key, f, encryptedPath); err != nil {
			log.Fatalf("Failed to encrypt file %s: %v", f, err)
		}
	}

	// Write the salt somewhere in the repo (non-secret):
	if err := writeSaltFile(filepath.Join(*folderPath, ".salt"), salt); err != nil {
		log.Fatalf("Failed to write salt file: %v", err)
	}

	if err := gitutils.AddAndCommit(config.RepoPath, *message); err != nil {
		fmt.Printf("Git commit failed (maybe nothing to commit?): %v\n", err)
	}
}

func writeSaltFile(path string, salt []byte) error {
	return fileutils.WriteFileAtomic(path, salt, 0644)
}
