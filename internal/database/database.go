/*
Copyright © 2025 Erik Wright <yo@oheriko.com>
*/
package database

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oheriko/tira/internal/config"
)

// Package represents an installed package
type Package struct {
	Name        string            `json:"name"`         // Derived from URL (e.g., "ollama", "docker")
	URL         string            `json:"url"`          // Original install URL
	ScriptHash  string            `json:"script_hash"`  // SHA256 of the script content
	Version     string            `json:"version"`      // Version detected or derived
	InstalledAt time.Time         `json:"installed_at"` // When it was installed
	InstalledBy string            `json:"installed_by"` // Tira version that installed it
	Metadata    PackageMetadata   `json:"metadata"`     // Additional metadata
	Files       []string          `json:"files"`        // Files created/modified
	Services    []string          `json:"services"`     // Services created
	Environment map[string]string `json:"environment"`  // Environment variables set
	Rollbacks   []RollbackEntry   `json:"rollbacks"`    // Previous versions for rollback
}

// PackageMetadata contains additional package information
type PackageMetadata struct {
	Description   string            `json:"description,omitempty"`   // Package description
	Homepage      string            `json:"homepage,omitempty"`      // Project homepage
	Documentation string            `json:"documentation,omitempty"` // Documentation URL
	Repository    string            `json:"repository,omitempty"`    // Source repository
	License       string            `json:"license,omitempty"`       // License
	Author        string            `json:"author,omitempty"`        // Author/maintainer
	Tags          []string          `json:"tags,omitempty"`          // Tags (e.g., "development", "gpu")
	Dependencies  []string          `json:"dependencies,omitempty"`  // System dependencies
	Conflicts     []string          `json:"conflicts,omitempty"`     // Conflicting packages
	Custom        map[string]string `json:"custom,omitempty"`        // Custom metadata
}

// RollbackEntry represents a previous version for rollback
type RollbackEntry struct {
	ScriptHash  string    `json:"script_hash"`  // Previous script hash
	Version     string    `json:"version"`      // Previous version
	InstalledAt time.Time `json:"installed_at"` // When this version was installed
	ScriptPath  string    `json:"script_path"`  // Path to cached script
	Files       []string  `json:"files"`        // Files that were modified
	Services    []string  `json:"services"`     // Services that were created
}

// Database manages the package database
type Database struct {
	dataDir string
}

// NewDatabase creates a new database instance
func NewDatabase() (*Database, error) {
	dataDir, err := config.DataPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get data path: %w", err)
	}

	return &Database{
		dataDir: dataDir,
	}, nil
}

// packagesDir returns the packages directory path
func (db *Database) packagesDir() string {
	return filepath.Join(db.dataDir, "packages")
}

// packagePath returns the path to a package's JSON file
func (db *Database) packagePath(name string) string {
	return filepath.Join(db.packagesDir(), fmt.Sprintf("%s.json", name))
}

// scriptsDir returns the scripts cache directory path
func (db *Database) scriptsDir() string {
	return filepath.Join(db.dataDir, "scripts")
}

// scriptPath returns the path to a cached script
func (db *Database) scriptPath(hash string) string {
	return filepath.Join(db.scriptsDir(), fmt.Sprintf("%s.sh", hash))
}

// EnsureDirectories creates necessary database directories
func (db *Database) EnsureDirectories() error {
	dirs := []string{
		db.packagesDir(),
		db.scriptsDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// SavePackage saves a package to the database
func (db *Database) SavePackage(pkg *Package) error {
	if err := db.EnsureDirectories(); err != nil {
		return err
	}

	packagePath := db.packagePath(pkg.Name)

	// Read existing package to preserve rollback history
	if existing, err := db.GetPackage(pkg.Name); err == nil {
		// Add current version to rollback history before updating
		rollback := RollbackEntry{
			ScriptHash:  existing.ScriptHash,
			Version:     existing.Version,
			InstalledAt: existing.InstalledAt,
			ScriptPath:  db.scriptPath(existing.ScriptHash),
			Files:       existing.Files,
			Services:    existing.Services,
		}

		// Prepend to rollback list (newest first)
		pkg.Rollbacks = append([]RollbackEntry{rollback}, existing.Rollbacks...)

		// Limit rollback history (configurable, default 5)
		maxRollbacks := 5 // TODO: Get from config
		if len(pkg.Rollbacks) > maxRollbacks {
			pkg.Rollbacks = pkg.Rollbacks[:maxRollbacks]
		}
	}

	// Write package file
	file, err := os.Create(packagePath)
	if err != nil {
		return fmt.Errorf("failed to create package file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(pkg); err != nil {
		return fmt.Errorf("failed to encode package: %w", err)
	}

	return nil
}

// GetPackage retrieves a package from the database
func (db *Database) GetPackage(name string) (*Package, error) {
	packagePath := db.packagePath(name)

	file, err := os.Open(packagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("package %s not found", name)
		}
		return nil, fmt.Errorf("failed to open package file: %w", err)
	}
	defer file.Close()

	var pkg Package
	if err := json.NewDecoder(file).Decode(&pkg); err != nil {
		return nil, fmt.Errorf("failed to decode package: %w", err)
	}

	return &pkg, nil
}

// ListPackages returns all installed packages
func (db *Database) ListPackages() ([]*Package, error) {
	packagesDir := db.packagesDir()

	// Check if packages directory exists
	if _, err := os.Stat(packagesDir); os.IsNotExist(err) {
		return []*Package{}, nil // No packages installed yet
	}

	entries, err := os.ReadDir(packagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read packages directory: %w", err)
	}

	var packages []*Package
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			name := strings.TrimSuffix(entry.Name(), ".json")
			pkg, err := db.GetPackage(name)
			if err != nil {
				// Log error but continue with other packages
				fmt.Fprintf(os.Stderr, "Warning: failed to load package %s: %v\n", name, err)
				continue
			}
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// DeletePackage removes a package from the database
func (db *Database) DeletePackage(name string) error {
	packagePath := db.packagePath(name)

	// Check if package exists
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return fmt.Errorf("package %s not found", name)
	}

	// Remove package file
	if err := os.Remove(packagePath); err != nil {
		return fmt.Errorf("failed to remove package file: %w", err)
	}

	return nil
}

// SaveScript saves a script to the cache and returns its hash
func (db *Database) SaveScript(content []byte) (string, error) {
	if err := db.EnsureDirectories(); err != nil {
		return "", err
	}

	// Calculate hash
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	scriptPath := db.scriptPath(hash)

	// Don't overwrite if already exists
	if _, err := os.Stat(scriptPath); err == nil {
		return hash, nil
	}

	// Write script file
	if err := os.WriteFile(scriptPath, content, 0755); err != nil {
		return "", fmt.Errorf("failed to write script: %w", err)
	}

	return hash, nil
}

// GetScript retrieves a cached script by hash
func (db *Database) GetScript(hash string) ([]byte, error) {
	scriptPath := db.scriptPath(hash)

	content, err := os.ReadFile(scriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("script with hash %s not found", hash)
		}
		return nil, fmt.Errorf("failed to read script: %w", err)
	}

	return content, nil
}

// PackageExists checks if a package is installed
func (db *Database) PackageExists(name string) bool {
	_, err := db.GetPackage(name)
	return err == nil
}

// GetPackageByURL finds a package by its original URL
func (db *Database) GetPackageByURL(url string) (*Package, error) {
	packages, err := db.ListPackages()
	if err != nil {
		return nil, err
	}

	for _, pkg := range packages {
		if pkg.URL == url {
			return pkg, nil
		}
	}

	return nil, fmt.Errorf("no package found for URL: %s", url)
}

// CleanupOldScripts removes unused script files
func (db *Database) CleanupOldScripts() error {
	packages, err := db.ListPackages()
	if err != nil {
		return err
	}

	// Collect all referenced script hashes
	referencedHashes := make(map[string]bool)
	for _, pkg := range packages {
		referencedHashes[pkg.ScriptHash] = true
		for _, rollback := range pkg.Rollbacks {
			referencedHashes[rollback.ScriptHash] = true
		}
	}

	// Walk scripts directory and remove unreferenced files
	scriptsDir := db.scriptsDir()
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		return nil // No scripts directory exists yet
	}

	return filepath.WalkDir(scriptsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".sh") {
			return nil
		}

		// Extract hash from filename
		hash := strings.TrimSuffix(d.Name(), ".sh")

		if !referencedHashes[hash] {
			fmt.Printf("Removing unused script: %s\n", hash[:8])
			return os.Remove(path)
		}

		return nil
	})
}

// GetStats returns database statistics
func (db *Database) GetStats() (map[string]interface{}, error) {
	packages, err := db.ListPackages()
	if err != nil {
		return nil, err
	}

	// Count script files
	scriptsDir := db.scriptsDir()
	scriptCount := 0
	if entries, err := os.ReadDir(scriptsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sh") {
				scriptCount++
			}
		}
	}

	// Calculate total rollback entries
	rollbackCount := 0
	for _, pkg := range packages {
		rollbackCount += len(pkg.Rollbacks)
	}

	stats := map[string]interface{}{
		"packages":        len(packages),
		"cached_scripts":  scriptCount,
		"rollback_points": rollbackCount,
		"database_path":   db.dataDir,
	}

	return stats, nil
}

// PackageNameFromURL derives a package name from a URL
func PackageNameFromURL(url string) string {
	// Simple URL-to-name conversion
	// https://ollama.com/install.sh -> ollama
	// https://get.docker.com -> docker
	// https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh -> nvm

	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Handle common patterns
	if strings.Contains(url, "ollama") {
		return "ollama"
	}
	if strings.Contains(url, "docker") {
		return "docker"
	}
	if strings.Contains(url, "nvm") {
		return "nvm"
	}
	if strings.Contains(url, "nixos.org") {
		return "nix"
	}
	if strings.Contains(url, "rustup") {
		return "rust"
	}
	if strings.Contains(url, "deno") {
		return "deno"
	}

	// Fallback: use domain name
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		domain := strings.Split(parts[0], ".")[0]
		return domain
	}

	return "unknown"
}
