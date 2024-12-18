package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var (
	ErrNoPassword  = errors.New("no encryption password provided")
	ErrNoRepoPath  = errors.New("no repository path provided")
	ErrNoWatchPath = errors.New("no watch path provided")
)

type Config struct {
	Password  string
	RepoPath  string
	WatchPath string
	RemoteURL string
}

// LoadConfig attempts to load configuration from various sources.
// Priority (lowest to highest):
// 1. Config file (config.yaml)
// 2. Environment variables
// If the password is still not set after these sources, you can consider prompting the user.
func LoadConfig() (*Config, error) {
	// Tell viper the name of the config file (without extension)
	viper.SetConfigName("config")
	// Set the path to look for the config file
	viper.AddConfigPath(".")
	// Optionally add more config paths if needed:
	// viper.AddConfigPath("/etc/git-fs/")

	// Set environment variable prefix if desired
	// This means if you have `ENCRYPTION_PASSWORD` in env, it maps to `password`
	viper.SetEnvPrefix("GITFS")
	viper.AutomaticEnv()

	// Try reading config file
	err := viper.ReadInConfig()
	if err != nil {
		// If config file not found, that's fine, we can proceed with defaults & env vars
		// If you want to fail if config file is missing, handle error here.
		fmt.Printf("No config file found, proceeding with defaults and environment variables.\n")
	}

	// Now extract values into our Config struct
	cfg := &Config{
		Password:  viper.GetString("password"),
		RepoPath:  viper.GetString("repo_path"),
		WatchPath: viper.GetString("watch_path"),
		RemoteURL: viper.GetString("remote_url"),
	}

	// Validate required fields
	if cfg.Password == "" {
		// Consider prompting user here if no password is provided:
		password, err := promptForPassword()
		if err != nil {
			return nil, ErrNoPassword
		}
		cfg.Password = password

		// For now, just return an error if no password is set
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

// Optional: If you want to prompt the user for a password when it's not provided
func promptForPassword() (string, error) {
	fmt.Print("Enter encryption password: ")
	var pw string
	_, err := fmt.Fscanln(os.Stdin, &pw)
	if err != nil {
		return "", err
	}
	if pw == "" {
		return "", ErrNoPassword
	}
	return pw, nil
}
