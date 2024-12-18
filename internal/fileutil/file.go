package fileutils

import (
	"git-fs/internal/crypto"
	"io/fs"
	"os"
	"path/filepath"
)

// GetFiles recursively fetches all files (not directories) from a given path.
func GetFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// Make sure a directory exists for output:
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// SafeStat is a helper to stat a file and avoid panics or confusion
func SafeStat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// FileExists checks if a file exists.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err == nil && !info.IsDir() {
		return true
	}
	return false
}

// ReadOrCreateSalt reads the salt if it exists or creates a new one.
func ReadOrCreateSalt(path string) ([]byte, error) {
	if FileExists(path) {
		return os.ReadFile(path)
	}
	salt, err := crypto.GenerateSalt() // from crypto key.go logic
	if err != nil {
		return nil, err
	}
	err = WriteFileAtomic(path, salt, 0644)
	return salt, err
}
