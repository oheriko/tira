/*
Copyright © 2025 Erik Wright <yo@oheriko.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/oheriko/tira/internal/config"
	"github.com/spf13/cobra"
)

var version = "dev" // Set by build process

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tira",
	Short: "Universal package manager for the curl | bash era",
	Long: `Tira 👑
Universal package manager for the curl | bash era

Tira brings version control and rollback capabilities to curl | bash
installation scripts. Install Ollama, Docker, NVM, Nix, or any
script-based tool with confidence, knowing you can always roll back
if something breaks.

Rule your installs 👑`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Ensure directories exist before running any commands
	if err := config.EnsureDirectories(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directories: %v\n", err)
		os.Exit(1)
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags that apply to all commands
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().String("config", "", "config file (default is $XDG_CONFIG_HOME/tira/config.toml)")
}
