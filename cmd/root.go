package cmd

import (
	"os"

	"git-fs/internal/logging"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "git-fs",
	Short: "A tool for git-backed encrypted cloud storage",
	Long: `git-fs watches a directory, encrypts its files, and stores them in a git repository.

This allows for secure, version-controlled backups.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logging.Logger.Error("Failed to execute root command", zap.Error(err))
		// Provide actionable advice if needed:
		// For example: "Please run `git-fs --help` for usage."
		os.Exit(1)
	}
}

func init() {
	// Add a --config flag to specify a config file
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: config.yaml)")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		// User specified a config file
		viper.SetConfigFile(cfgFile)
	} else {
		// Default to config.yaml in current directory
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
	}

	// Allow environment variable overrides with prefix GITFS_
	viper.SetEnvPrefix("GITFS")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logging.Logger.Info("Using config file", zap.String("file", viper.ConfigFileUsed()))
	} else {
		// It's okay if no config file is found; environment variables may suffice.
		// If you want to handle this as a warning or provide guidance:
		logging.Logger.Debug("No config file found, using defaults and environment variables")
	}
}
