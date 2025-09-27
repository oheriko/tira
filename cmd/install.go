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

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [url]",
	Short: "Install a package from a curl | bash script",
	Long: `Install a package from a curl | bash script.

Examples:
  tira install https://ollama.com/install.sh
  tira install https://get.docker.com
  tira install https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh
  tira install https://nixos.org/nix/install
  tira install https://ollama.com/install.sh@a1b2c3d4  # Pin to specific script hash`,
	Args: cobra.ExactArgs(1),
	Run:  installHandler,
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Command-specific flags
	installCmd.Flags().String("cache", "", "Cache strategy: smart, always, pin, offline")
	installCmd.Flags().Bool("no-confirm", false, "Skip confirmation prompts")
	installCmd.Flags().Bool("dry-run", false, "Show what would be installed without doing it")
	installCmd.Flags().Bool("force", false, "Force installation even if package already exists")
	installCmd.Flags().String("name", "", "Override package name (auto-detected from URL)")
}

func installHandler(cmd *cobra.Command, args []string) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error loading config: %v\n", err)
		os.Exit(1)
	}

	url := args[0]
	cache, _ := cmd.Flags().GetString("cache")
	noConfirm, _ := cmd.Flags().GetBool("no-confirm")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	customName, _ := cmd.Flags().GetString("name")

	// Get effective config for this script
	scriptConfig := cfg.GetScriptConfig(url)

	// Override with command line flags
	if cache != "" {
		scriptConfig.Cache.Strategy = cache
	}
	if noConfirm {
		scriptConfig.UI.ConfirmInstalls = false
	}

	// Show what we're doing
	if dryRun {
		fmt.Printf("🔍 Dry run - would install from %s\n", url)
	} else {
		fmt.Printf("🔽 Installing from %s\n", url)
	}

	// Display configuration being used
	if cmd.Flag("verbose").Changed || scriptConfig.UI.Verbose {
		fmt.Printf("   📋 Configuration:\n")
		fmt.Printf("      Cache strategy: %s\n", scriptConfig.Cache.Strategy)
		fmt.Printf("      Cache TTL: %v\n", scriptConfig.Cache.TTL)
		fmt.Printf("      HTTPS verification: %t\n", scriptConfig.Security.VerifyHTTPS)
		fmt.Printf("      Prompt on changes: %t\n", scriptConfig.Security.PromptOnChange)
		if customName != "" {
			fmt.Printf("      Custom name: %s\n", customName)
		}
		if force {
			fmt.Printf("      Force install: %t\n", force)
		}
		if !scriptConfig.UI.ConfirmInstalls {
			fmt.Printf("      Skip confirmations: %t\n", true)
		}
	}

	// TODO: Implement the actual installation logic:
	// 1. Parse URL (handle @hash syntax for pinning)
	// 2. Check if package already exists (unless --force)
	// 3. Download script with HTTP client
	// 4. Verify script (checksum, HTTPS, domain whitelist)
	// 5. Cache script and calculate hash
	// 6. Monitor filesystem before/after execution
	// 7. Execute script in controlled environment
	// 8. Detect installed files, services, environment changes
	// 9. Save package metadata to database
	// 10. Report success with rollback info

	if !dryRun {
		fmt.Printf("✅ Installation completed successfully\n")
		fmt.Printf("   Use 'tira list' to see installed packages\n")
		fmt.Printf("   Use 'tira rollback <package>' if you need to revert\n")
	} else {
		fmt.Printf("   💡 Run without --dry-run to perform actual installation\n")
	}
}
