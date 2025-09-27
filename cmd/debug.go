/*
Copyright © 2025 Erik Wright <yo@oheriko.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oheriko/tira/internal/config"
	"github.com/oheriko/tira/internal/database"
	"github.com/spf13/cobra"
)

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:    "debug",
	Short:  "Debug commands (for development)",
	Long:   `Debug and development commands for testing Tira functionality.`,
	Hidden: true, // Hide from normal help output
}

var debugAddCmd = &cobra.Command{
	Use:   "add-test-package",
	Short: "Add test packages to the database",
	Long:  `Add sample test packages to the database for testing list and rollback functionality.`,
	Run:   debugAddHandler,
}

var debugStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show database and system statistics",
	Long:  `Display detailed statistics about the package database, cache usage, and system paths.`,
	Run:   debugStatsHandler,
}

var debugCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up unused cached scripts",
	Long:  `Remove script files that are no longer referenced by any installed packages.`,
	Run:   debugCleanupHandler,
}

var debugPathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Show all Tira directory paths",
	Long:  `Display the paths used by Tira for configuration, data, and cache storage.`,
	Run:   debugPathsHandler,
}

var debugResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset database (removes all packages)",
	Long:  `Remove all installed packages from the database. This does NOT uninstall the actual software.`,
	Run:   debugResetHandler,
}

func init() {
	rootCmd.AddCommand(debugCmd)

	// Add debug subcommands
	debugCmd.AddCommand(debugAddCmd)
	debugCmd.AddCommand(debugStatsCmd)
	debugCmd.AddCommand(debugCleanupCmd)
	debugCmd.AddCommand(debugPathsCmd)
	debugCmd.AddCommand(debugResetCmd)

	// Add flags
	debugResetCmd.Flags().Bool("confirm", false, "Confirm the reset operation")
}

func debugAddHandler(cmd *cobra.Command, args []string) {
	db, err := database.NewDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🔧 Adding test packages to database...")

	// Test package 1: Ollama
	pkg1 := &database.Package{
		Name:        "ollama",
		URL:         "https://ollama.com/install.sh",
		ScriptHash:  "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456",
		Version:     "v0.1.17",
		InstalledAt: time.Now().Add(-2 * 24 * time.Hour), // 2 days ago
		InstalledBy: "tira/dev",
		Metadata: database.PackageMetadata{
			Description:  "Large language model runner",
			Homepage:     "https://ollama.com",
			Repository:   "https://github.com/ollama/ollama",
			License:      "MIT",
			Tags:         []string{"ai", "gpu", "development", "llm"},
			Dependencies: []string{"nvidia-docker", "curl"},
		},
		Files:    []string{"/usr/local/bin/ollama", "/etc/systemd/system/ollama.service"},
		Services: []string{"ollama"},
		Environment: map[string]string{
			"OLLAMA_HOST": "0.0.0.0",
			"OLLAMA_PORT": "11434",
		},
		Rollbacks: []database.RollbackEntry{
			{
				ScriptHash:  "b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456a1",
				Version:     "v0.1.16",
				InstalledAt: time.Now().Add(-5 * 24 * time.Hour),
				ScriptPath:  "/cached/script/path",
				Files:       []string{"/usr/local/bin/ollama"},
				Services:    []string{"ollama"},
			},
		},
	}

	if err := db.SavePackage(pkg1); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error saving ollama package: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("   ✅ Added ollama v0.1.17")

	// Test package 2: Docker
	pkg2 := &database.Package{
		Name:        "docker",
		URL:         "https://get.docker.com",
		ScriptHash:  "c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456a1b2",
		Version:     "v24.0.7",
		InstalledAt: time.Now().Add(-7 * 24 * time.Hour), // 1 week ago
		InstalledBy: "tira/dev",
		Metadata: database.PackageMetadata{
			Description:  "Container runtime platform",
			Homepage:     "https://docker.com",
			Repository:   "https://github.com/docker/docker-ce",
			License:      "Apache-2.0",
			Tags:         []string{"containers", "development", "devops"},
			Dependencies: []string{"systemd"},
		},
		Files:    []string{"/usr/bin/docker", "/usr/bin/dockerd", "/etc/docker/daemon.json"},
		Services: []string{"docker"},
		Environment: map[string]string{
			"DOCKER_HOST": "unix:///var/run/docker.sock",
		},
		Rollbacks: []database.RollbackEntry{
			{
				ScriptHash:  "d4e5f6789012345678901234567890abcdef1234567890abcdef123456a1b2c3",
				Version:     "v24.0.6",
				InstalledAt: time.Now().Add(-14 * 24 * time.Hour),
				ScriptPath:  "/cached/script/path2",
				Files:       []string{"/usr/bin/docker", "/usr/bin/dockerd"},
				Services:    []string{"docker"},
			},
			{
				ScriptHash:  "e5f6789012345678901234567890abcdef1234567890abcdef123456a1b2c3d4",
				Version:     "v24.0.5",
				InstalledAt: time.Now().Add(-21 * 24 * time.Hour),
				ScriptPath:  "/cached/script/path3",
				Files:       []string{"/usr/bin/docker"},
				Services:    []string{"docker"},
			},
		},
	}

	if err := db.SavePackage(pkg2); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error saving docker package: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("   ✅ Added docker v24.0.7")

	// Test package 3: NVM
	pkg3 := &database.Package{
		Name:        "nvm",
		URL:         "https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh",
		ScriptHash:  "f6789012345678901234567890abcdef1234567890abcdef123456a1b2c3d4e5",
		Version:     "v0.40.3",
		InstalledAt: time.Now().Add(-1 * time.Hour), // 1 hour ago
		InstalledBy: "tira/dev",
		Metadata: database.PackageMetadata{
			Description:  "Node Version Manager",
			Homepage:     "https://github.com/nvm-sh/nvm",
			Repository:   "https://github.com/nvm-sh/nvm",
			License:      "MIT",
			Tags:         []string{"nodejs", "development", "version-manager"},
			Dependencies: []string{"curl", "bash"},
		},
		Files: []string{
			"$HOME/.nvm/nvm.sh",
			"$HOME/.bashrc", // modified
			"$HOME/.zshrc",  // modified
		},
		Services: []string{}, // No services for NVM
		Environment: map[string]string{
			"NVM_DIR": "$HOME/.nvm",
		},
		Rollbacks: []database.RollbackEntry{}, // Fresh install
	}

	if err := db.SavePackage(pkg3); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error saving nvm package: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("   ✅ Added nvm v0.40.3")

	fmt.Println("\n🎉 Test packages added successfully!")
	fmt.Println("💡 Try: tira list --show-files --show-rollbacks")
}

func debugStatsHandler(cmd *cobra.Command, args []string) {
	db, err := database.NewDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}

	stats, err := db.GetStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error getting stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("📊 Database Statistics:")
	for key, value := range stats {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// Additional system info
	fmt.Println("\n🖥️  System Information:")

	configPath, _ := config.ConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		if info, err := os.Stat(configPath); err == nil {
			fmt.Printf("   config_file_size: %d bytes\n", info.Size())
			fmt.Printf("   config_modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
		}
	}

	dataPath, _ := config.DataPath()
	if _, err := os.Stat(dataPath); err == nil {
		fmt.Printf("   data_dir_size: calculating...\n")
		if size, err := getDirSize(dataPath); err == nil {
			fmt.Printf("   data_dir_size: %d bytes\n", size)
		}
	}

	cachePath, _ := config.CachePath()
	if _, err := os.Stat(cachePath); err == nil {
		if size, err := getDirSize(cachePath); err == nil {
			fmt.Printf("   cache_dir_size: %d bytes\n", size)
		}
	}
}

func debugCleanupHandler(cmd *cobra.Command, args []string) {
	db, err := database.NewDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🧹 Cleaning up unused cached scripts...")

	if err := db.CleanupOldScripts(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error during cleanup: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Cleanup completed")
}

func debugPathsHandler(cmd *cobra.Command, args []string) {
	fmt.Println("📁 Tira Directory Paths:")

	configPath, err := config.ConfigPath()
	if err != nil {
		fmt.Printf("   ❌ Config file: error - %v\n", err)
	} else {
		fmt.Printf("   📋 Config file: %s\n", configPath)
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("      ✅ exists\n")
		} else {
			fmt.Printf("      ❌ does not exist\n")
		}
	}

	dataPath, err := config.DataPath()
	if err != nil {
		fmt.Printf("   ❌ Data directory: error - %v\n", err)
	} else {
		fmt.Printf("   💾 Data directory: %s\n", dataPath)
		if _, err := os.Stat(dataPath); err == nil {
			fmt.Printf("      ✅ exists\n")
		} else {
			fmt.Printf("      ❌ does not exist\n")
		}
	}

	cachePath, err := config.CachePath()
	if err != nil {
		fmt.Printf("   ❌ Cache directory: error - %v\n", err)
	} else {
		fmt.Printf("   🗂️  Cache directory: %s\n", cachePath)
		if _, err := os.Stat(cachePath); err == nil {
			fmt.Printf("      ✅ exists\n")
		} else {
			fmt.Printf("      ❌ does not exist\n")
		}
	}

	// Show subdirectories
	fmt.Println("\n📂 Subdirectories:")
	if dataPath, err := config.DataPath(); err == nil {
		packageDir := filepath.Join(dataPath, "packages")
		scriptsDir := filepath.Join(dataPath, "scripts")

		fmt.Printf("   📦 Packages: %s\n", packageDir)
		if _, err := os.Stat(packageDir); err == nil {
			fmt.Printf("      ✅ exists\n")
		} else {
			fmt.Printf("      ❌ does not exist\n")
		}

		fmt.Printf("   📜 Scripts: %s\n", scriptsDir)
		if _, err := os.Stat(scriptsDir); err == nil {
			fmt.Printf("      ✅ exists\n")
		} else {
			fmt.Printf("      ❌ does not exist\n")
		}
	}
}

func debugResetHandler(cmd *cobra.Command, args []string) {
	confirm, _ := cmd.Flags().GetBool("confirm")

	if !confirm {
		fmt.Println("⚠️  This will remove ALL installed packages from the database.")
		fmt.Println("   This does NOT uninstall the actual software from your system.")
		fmt.Println("   To confirm, run: tira debug reset --confirm")
		return
	}

	db, err := database.NewDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}

	packages, err := db.ListPackages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error listing packages: %v\n", err)
		os.Exit(1)
	}

	if len(packages) == 0 {
		fmt.Println("📦 No packages to remove")
		return
	}

	fmt.Printf("🗑️  Removing %d packages from database...\n", len(packages))

	for _, pkg := range packages {
		if err := db.DeletePackage(pkg.Name); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error removing %s: %v\n", pkg.Name, err)
		} else {
			fmt.Printf("   ✅ Removed %s\n", pkg.Name)
		}
	}

	fmt.Println("✅ Database reset completed")
	fmt.Println("💡 The actual software is still installed on your system")
}

func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
