package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "git-fs",
	Short: "A tool for git-backed encrypted cloud storage",
	Long: `git-fs watches a directory, encrypts its files, and stores them in a git repository.

This allows for secure, version-controlled backups.`}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() { // Add a --config flag to specify a config file
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: config.yaml)")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		// User specified a config file
		viper.SetConfigFile(cfgFile)
	} else { // Default to config.yaml in current directory
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
	}

	// Allow environment variable overrides with prefix GITFS_
	viper.SetEnvPrefix("GITFS")
	viper.AutomaticEnv()

	// If a config file is found, read it
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

}
