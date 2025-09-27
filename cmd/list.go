/*
Copyright © 2025 Erik Wright <yo@oheriko.com>
*/
package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/oheriko/tira/internal/config"
	"github.com/oheriko/tira/internal/database"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed packages",
	Long: `List all packages installed via Tira with their versions and metadata.

Examples:
  tira list                    # List all packages
  tira list --show-hashes      # Include script hashes
  tira list --show-files       # Show installed files
  tira list --show-services    # Show installed services
  tira list --verbose          # Show detailed information`,
	Aliases: []string{"ls"},
	Run:     listHandler,
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Display options
	listCmd.Flags().Bool("show-hashes", false, "Show script hashes")
	listCmd.Flags().Bool("show-files", false, "Show installed files")
	listCmd.Flags().Bool("show-services", false, "Show installed services")
	listCmd.Flags().Bool("show-env", false, "Show environment variables")
	listCmd.Flags().Bool("show-rollbacks", false, "Show available rollback points")

	// Filtering options
	listCmd.Flags().StringSlice("tags", []string{}, "Filter by tags (e.g., --tags=gpu,ai)")
	listCmd.Flags().String("sort", "name", "Sort by: name, date, version")
	listCmd.Flags().Bool("reverse", false, "Reverse sort order")
}

func listHandler(cmd *cobra.Command, args []string) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := database.NewDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}

	// Get flag values
	showHashes, _ := cmd.Flags().GetBool("show-hashes")
	showFiles, _ := cmd.Flags().GetBool("show-files")
	showServices, _ := cmd.Flags().GetBool("show-services")
	showEnv, _ := cmd.Flags().GetBool("show-env")
	showRollbacks, _ := cmd.Flags().GetBool("show-rollbacks")
	tagFilter, _ := cmd.Flags().GetStringSlice("tags")
	sortBy, _ := cmd.Flags().GetString("sort")
	reverse, _ := cmd.Flags().GetBool("reverse")

	// Use config defaults if flags not explicitly set
	if !cmd.Flags().Changed("show-hashes") {
		showHashes = cfg.UI.ShowHashes
	}

	packages, err := db.ListPackages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error listing packages: %v\n", err)
		os.Exit(1)
	}

	if len(packages) == 0 {
		fmt.Println("📦 No packages installed yet")
		fmt.Println()
		fmt.Println("💡 Get started with:")
		fmt.Println("   tira install https://ollama.com/install.sh")
		fmt.Println("   tira install https://get.docker.com")
		fmt.Println("   tira install https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh")
		return
	}

	// Filter packages by tags if specified
	if len(tagFilter) > 0 {
		packages = filterPackagesByTags(packages, tagFilter)
		if len(packages) == 0 {
			fmt.Printf("📦 No packages found with tags: %s\n", strings.Join(tagFilter, ", "))
			return
		}
	}

	// Sort packages
	sortPackages(packages, sortBy, reverse)

	// Display header
	fmt.Printf("📦 Installed packages (%d):\n", len(packages))
	if len(tagFilter) > 0 {
		fmt.Printf("   🏷️  Filtered by tags: %s\n", strings.Join(tagFilter, ", "))
	}
	fmt.Println()

	// Display packages
	for i, pkg := range packages {
		displayPackage(pkg, showHashes, showFiles, showServices, showEnv, showRollbacks)

		// Add spacing between packages (except for the last one)
		if i < len(packages)-1 {
			fmt.Println()
		}
	}
}

func displayPackage(pkg *database.Package, showHashes, showFiles, showServices, showEnv, showRollbacks bool) {
	// Main package info
	if showHashes {
		fmt.Printf("  📦 %s %s (script hash: %s)\n",
			pkg.Name,
			pkg.Version,
			pkg.ScriptHash[:8]+"...")
	} else {
		fmt.Printf("  📦 %s %s\n", pkg.Name, pkg.Version)
	}

	// Description
	if pkg.Metadata.Description != "" {
		fmt.Printf("     %s\n", pkg.Metadata.Description)
	}

	// Installation info
	fmt.Printf("     🔗 %s\n", pkg.URL)
	fmt.Printf("     📅 Installed %s (%s)\n",
		formatTimeAgo(pkg.InstalledAt),
		pkg.InstalledAt.Format("2006-01-02 15:04"))

	// Tags
	if len(pkg.Metadata.Tags) > 0 {
		fmt.Printf("     🏷️  %s\n", strings.Join(pkg.Metadata.Tags, ", "))
	}

	// Files
	if showFiles && len(pkg.Files) > 0 {
		fmt.Printf("     📁 Files (%d):\n", len(pkg.Files))
		for _, file := range pkg.Files {
			fmt.Printf("        %s\n", file)
		}
	}

	// Services
	if showServices && len(pkg.Services) > 0 {
		fmt.Printf("     ⚙️  Services: %s\n", strings.Join(pkg.Services, ", "))
	}

	// Environment variables
	if showEnv && len(pkg.Environment) > 0 {
		fmt.Printf("     🌍 Environment:\n")
		for key, value := range pkg.Environment {
			fmt.Printf("        %s=%s\n", key, value)
		}
	}

	// Rollback info
	if showRollbacks && len(pkg.Rollbacks) > 0 {
		fmt.Printf("     🔄 Rollback points (%d):\n", len(pkg.Rollbacks))
		for i, rollback := range pkg.Rollbacks {
			if i >= 3 { // Limit to first 3 for readability
				fmt.Printf("        ... and %d more\n", len(pkg.Rollbacks)-3)
				break
			}
			fmt.Printf("        %s (%s) - %s\n",
				rollback.Version,
				rollback.ScriptHash[:8],
				formatTimeAgo(rollback.InstalledAt))
		}
	} else if len(pkg.Rollbacks) > 0 {
		fmt.Printf("     └─ %d rollback point(s) available\n", len(pkg.Rollbacks))
	}
}

func filterPackagesByTags(packages []*database.Package, tags []string) []*database.Package {
	var filtered []*database.Package

	for _, pkg := range packages {
		hasAllTags := true
		for _, requiredTag := range tags {
			found := false
			for _, pkgTag := range pkg.Metadata.Tags {
				if strings.EqualFold(pkgTag, requiredTag) {
					found = true
					break
				}
			}
			if !found {
				hasAllTags = false
				break
			}
		}
		if hasAllTags {
			filtered = append(filtered, pkg)
		}
	}

	return filtered
}

func sortPackages(packages []*database.Package, sortBy string, reverse bool) {
	switch sortBy {
	case "date":
		sort.Slice(packages, func(i, j int) bool {
			if reverse {
				return packages[i].InstalledAt.After(packages[j].InstalledAt)
			}
			return packages[i].InstalledAt.Before(packages[j].InstalledAt)
		})
	case "version":
		sort.Slice(packages, func(i, j int) bool {
			if reverse {
				return packages[i].Version > packages[j].Version
			}
			return packages[i].Version < packages[j].Version
		})
	default: // name
		sort.Slice(packages, func(i, j int) bool {
			if reverse {
				return packages[i].Name > packages[j].Name
			}
			return packages[i].Name < packages[j].Name
		})
	}
}

func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 30*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		return t.Format("Jan 2, 2006")
	}
}
