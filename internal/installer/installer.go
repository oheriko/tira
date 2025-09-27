/*
Copyright © 2025 Erik Wright <yo@oheriko.com>
*/
package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/oheriko/tira/internal/config"
	"github.com/oheriko/tira/internal/database"
)

// Installer handles package installation
type Installer struct {
	db     *database.Database
	config *config.Config
}

// NewInstaller creates a new installer instance
func NewInstaller(cfg *config.Config) (*Installer, error) {
	db, err := database.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	return &Installer{
		db:     db,
		config: cfg,
	}, nil
}

// InstallOptions contains installation options
type InstallOptions struct {
	URL        string
	CustomName string
	Force      bool
	DryRun     bool
}

// Install downloads and executes a script, tracking it in the database
func (i *Installer) Install(opts InstallOptions) error {
	fmt.Printf("🔽 Installing from %s\n", opts.URL)

	// 1. Derive package name
	packageName := opts.CustomName
	if packageName == "" {
		packageName = database.PackageNameFromURL(opts.URL)
	}

	// 2. Check if package already exists
	if !opts.Force && i.db.PackageExists(packageName) {
		return fmt.Errorf("package %s already installed (use --force to reinstall)", packageName)
	}

	// 3. Download script
	fmt.Printf("   📥 Downloading script...\n")
	scriptContent, err := i.downloadScript(opts.URL)
	if err != nil {
		return fmt.Errorf("failed to download script: %w", err)
	}

	// 4. Calculate hash and cache script
	fmt.Printf("   🔐 Calculating hash...\n")
	scriptHash, err := i.db.SaveScript(scriptContent)
	if err != nil {
		return fmt.Errorf("failed to cache script: %w", err)
	}

	fmt.Printf("   📋 Script hash: %s\n", scriptHash[:16]+"...")

	// 5. If dry run, stop here
	if opts.DryRun {
		fmt.Printf("   🔍 Dry run complete - script would be executed\n")
		return nil
	}

	// 6. Execute script
	fmt.Printf("   ⚡ Executing script...\n")
	version, err := i.executeScript(scriptContent)
	if err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	// 7. Save package to database
	fmt.Printf("   💾 Saving package metadata...\n")
	pkg := &database.Package{
		Name:        packageName,
		URL:         opts.URL,
		ScriptHash:  scriptHash,
		Version:     version,
		InstalledAt: time.Now(),
		InstalledBy: "tira/dev", // TODO: Get from build version
		Metadata:    i.detectMetadata(opts.URL, packageName),
		Files:       []string{}, // TODO: Will be populated with filesystem monitoring
		Services:    []string{}, // TODO: Will be populated with service detection
		Environment: make(map[string]string),
		Rollbacks:   []database.RollbackEntry{},
	}

	if err := i.db.SavePackage(pkg); err != nil {
		return fmt.Errorf("failed to save package: %w", err)
	}

	fmt.Printf("✅ Successfully installed %s %s\n", packageName, version)
	fmt.Printf("   Use 'tira list' to see installed packages\n")

	return nil
}

// downloadScript downloads a script from a URL
func (i *Installer) downloadScript(url string) ([]byte, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: i.config.Network.Timeout,
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent
	req.Header.Set("User-Agent", i.config.Network.UserAgent)

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check size limit
	if int64(len(content)) > i.config.Security.MaxScriptSize {
		return nil, fmt.Errorf("script too large: %d bytes (limit: %d)",
			len(content), i.config.Security.MaxScriptSize)
	}

	// Basic security check - ensure HTTPS if required
	if i.config.Security.VerifyHTTPS && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("HTTPS required but URL is not secure: %s", url)
	}

	return content, nil
}

// executeScript executes a script and returns the detected version
func (i *Installer) executeScript(scriptContent []byte) (string, error) {
	// Create temporary script file
	tmpFile, err := os.CreateTemp("", "tira-install-*.sh")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up

	// Write script content
	if _, err := tmpFile.Write(scriptContent); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write script: %w", err)
	}

	// Make executable
	if err := tmpFile.Chmod(0755); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to make script executable: %w", err)
	}
	tmpFile.Close()

	// Check if this is a self-installation to prevent recursion
	if strings.Contains(string(scriptContent), "tira install") &&
		strings.Contains(string(scriptContent), "tira.sh") {
		fmt.Printf("   🔄 Self-installation detected - executing in safe mode\n")

		// Set environment variable to prevent nested self-installs
		cmd := exec.Command("/bin/bash", tmpFile.Name())
		cmd.Env = append(os.Environ(), "TIRA_INSTALLING=true")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Don't inherit stdin for self-install scripts
		cmd.Stdin = nil

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("script execution failed: %w", err)
		}
	} else {
		// Normal script execution with full stdin/stdout
		cmd := exec.Command("/bin/bash", tmpFile.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("script execution failed: %w", err)
		}
	}

	// For now, return a simple version based on timestamp
	// TODO: Implement proper version detection
	return fmt.Sprintf("installed-%s", time.Now().Format("20060102-150405")), nil
}

// detectMetadata tries to detect package metadata from URL and name
func (i *Installer) detectMetadata(url, name string) database.PackageMetadata {
	metadata := database.PackageMetadata{
		Tags: []string{},
	}

	// Add basic metadata based on known packages
	switch name {
	case "ollama":
		metadata.Description = "Large language model runner"
		metadata.Homepage = "https://ollama.com"
		metadata.Repository = "https://github.com/ollama/ollama"
		metadata.License = "MIT"
		metadata.Tags = []string{"ai", "gpu", "development", "llm"}
	case "docker":
		metadata.Description = "Container runtime platform"
		metadata.Homepage = "https://docker.com"
		metadata.Repository = "https://github.com/docker/docker-ce"
		metadata.License = "Apache-2.0"
		metadata.Tags = []string{"containers", "development", "devops"}
	case "nvm":
		metadata.Description = "Node Version Manager"
		metadata.Homepage = "https://github.com/nvm-sh/nvm"
		metadata.Repository = "https://github.com/nvm-sh/nvm"
		metadata.License = "MIT"
		metadata.Tags = []string{"nodejs", "development", "version-manager"}
	case "nix":
		metadata.Description = "Nix package manager"
		metadata.Homepage = "https://nixos.org"
		metadata.Repository = "https://github.com/NixOS/nix"
		metadata.License = "LGPL-2.1"
		metadata.Tags = []string{"package-manager", "development", "functional"}
	case "rust":
		metadata.Description = "Rust programming language toolchain"
		metadata.Homepage = "https://rustup.rs"
		metadata.Repository = "https://github.com/rust-lang/rustup"
		metadata.License = "MIT"
		metadata.Tags = []string{"rust", "development", "toolchain"}
	default:
		metadata.Description = fmt.Sprintf("Package installed from %s", url)
		metadata.Tags = []string{"unknown"}
	}

	return metadata
}
