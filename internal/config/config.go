package config

import (
	"errors"
	"os"
)

var (
	ErrNoPassword  = errors.New("no encryption password provided")
	ErrNoRepoPath  = errors.New("no repository path provided")
	ErrNoWatchPath = errors.New("no watch path provided")
)

// Config holds our application configuration.
type Config struct {
	Password  string
	RepoPath  string
	WatchPath string
	RemoteURL string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Password:  os.Getenv("ENCRYPTION_PASSWORD"),
		RepoPath:  os.Getenv("REPO_PATH"),
		WatchPath: os.Getenv("WATCH_PATH"),
		RemoteURL: os.Getenv("REMOTE_REPO"), // optional
	}

	if cfg.Password == "" {
		return nil, ErrNoPassword
	}
	if cfg.RepoPath == "" {
		return nil, ErrNoRepoPath
	}
	if cfg.WatchPath == "" {
		return nil, ErrNoWatchPath
	}

	return cfg, nil
}
