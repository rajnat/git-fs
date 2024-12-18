package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"git-fs/internal/config"
	"git-fs/internal/crypto"
	fileutils "git-fs/internal/fileutil"
)

func main() {
	repoPath := flag.String("path", ".", "Path to the repo")
	flag.Parse()

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error getting password: %v", err)
	}
	saltPath := filepath.Join(*repoPath, ".salt")
	salt, err := os.ReadFile(saltPath)
	if err != nil {
		log.Fatalf("Failed to read salt: %v", err)
	}

	key, err := crypto.DeriveKey(config.Password, salt)
	if err != nil {
		log.Fatalf("Error deriving key: %v", err)
	}

	encryptedRoot := filepath.Join(*repoPath, ".encrypted")
	files, err := fileutils.GetFiles(encryptedRoot)
	if err != nil {
		log.Fatalf("Error getting encrypted files: %v", err)
	}

	for _, encFile := range files {
		rel, _ := filepath.Rel(encryptedRoot, encFile)
		// Remove .enc extension
		plaintextRel := strings.TrimSuffix(rel, ".enc")
		plaintextPath := filepath.Join(*repoPath, plaintextRel)
		if err := fileutils.EnsureDir(filepath.Dir(plaintextPath)); err != nil {
			log.Fatalf("Failed to ensure directory for %s: %v", plaintextPath, err)
		}
		if err := crypto.DecryptFile(key, encFile, plaintextPath); err != nil {
			log.Fatalf("Failed to decrypt file %s: %v", encFile, err)
		}
	}
}
