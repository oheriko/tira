/*
Copyright © 2025 Erik Wright <yo@oheriko.com>
*/
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
)

// Config represents the main configuration structure
type Config struct {
	Cache    CacheConfig    `toml:"cache"`
	Security SecurityConfig `toml:"security"`
	Rollback RollbackConfig `toml:"rollback"`
	UI       UIConfig       `toml:"ui"`
	Network  NetworkConfig  `toml:"network"`
}

// CacheConfig controls caching behavior
type CacheConfig struct {
	Strategy    string        `toml:"strategy"`     // "smart", "always", "pin", "offline"
	TTL         time.Duration `toml:"ttl"`          // How long to consider cache fresh
	AutoCleanup bool          `toml:"auto_cleanup"` // Clean old script versions automatically
}

// SecurityConfig controls security settings
type SecurityConfig struct {
	VerifyHTTPS     bool     `toml:"verify_https"`     // Require HTTPS for scripts
	PromptOnChange  bool     `toml:"prompt_on_change"` // Ask user when script hash changes
	BlockedDomains  []string `toml:"blocked_domains"`  // Domains to never install from
	AllowedDomains  []string `toml:"allowed_domains"`  // Only allow these domains (if set)
	MaxScriptSize   int64    `toml:"max_script_size"`  // Max script size in bytes
	RequireChecksum bool     `toml:"require_checksum"` // Require scripts to provide checksums
}

// RollbackConfig controls rollback behavior
type RollbackConfig struct {
	KeepVersions int  `toml:"keep_versions"` // How many old versions to keep
	AutoBackup   bool `toml:"auto_backup"`   // Always backup before install
}

// UIConfig controls user interface behavior
type UIConfig struct {
	ShowProgress    bool `toml:"show_progress"`    // Show progress bars
	ShowHashes      bool `toml:"show_hashes"`      // Show hashes in list by default
	ConfirmInstalls bool `toml:"confirm_installs"` // Skip "are you sure?" prompts
	UseColor        bool `toml:"use_color"`        // Use colored output
	Verbose         bool `toml:"verbose"`          // Verbose output by default
}

// NetworkConfig controls network behavior
type NetworkConfig struct {
	Timeout    time.Duration `toml:"timeout"`     // HTTP timeout
	Retries    int           `toml:"retries"`     // Number of retries
	UserAgent  string        `toml:"user_agent"`  // HTTP User-Agent header
	Proxy      string        `toml:"proxy"`       // HTTP proxy URL
	SkipVerify bool          `toml:"skip_verify"` // Skip TLS verification (dangerous)
}

// ScriptOverride allows per-script configuration
type ScriptOverride struct {
	Cache    CacheConfig    `toml:"cache"`
	Security SecurityConfig `toml:"security"`
	Rollback RollbackConfig `toml:"rollback"`
}

// ConfigWithOverrides includes script-specific overrides
type ConfigWithOverrides struct {
	Config
	Scripts map[string]ScriptOverride `toml:"scripts"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Cache: CacheConfig{
			Strategy:    "smart",
			TTL:         24 * time.Hour,
			AutoCleanup: true,
		},
		Security: SecurityConfig{
			VerifyHTTPS:     true,
			PromptOnChange:  true,
			BlockedDomains:  []string{},
			AllowedDomains:  []string{},
			MaxScriptSize:   10 * 1024 * 1024, // 10MB
			RequireChecksum: false,
		},
		Rollback: RollbackConfig{
			KeepVersions: 5,
			AutoBackup:   true,
		},
		UI: UIConfig{
			ShowProgress:    true,
			ShowHashes:      false,
			ConfirmInstalls: false,
			UseColor:        true,
			Verbose:         false,
		},
		Network: NetworkConfig{
			Timeout:    30 * time.Second,
			Retries:    3,
			UserAgent:  "tira/1.0.0",
			Proxy:      "",
			SkipVerify: false,
		},
	}
}

// ConfigPath returns the full path to the config file
func ConfigPath() (string, error) {
	configDir := filepath.Join(xdg.ConfigHome, "tira")
	return filepath.Join(configDir, "config.toml"), nil
}

// DataPath returns the full path to the data directory
func DataPath() (string, error) {
	return filepath.Join(xdg.DataHome, "tira"), nil
}

// CachePath returns the full path to the cache directory
func CachePath() (string, error) {
	return filepath.Join(xdg.CacheHome, "tira"), nil
}

// Load loads configuration from the config file
func Load() (*ConfigWithOverrides, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Start with default config
	config := &ConfigWithOverrides{
		Config:  *DefaultConfig(),
		Scripts: make(map[string]ScriptOverride),
	}

	// If config file doesn't exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	// Read and parse config file
	if _, err := toml.DecodeFile(configPath, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Save saves the configuration to the config file
func (c *ConfigWithOverrides) Save() error {
	configPath, err := ConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create config file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Write config as TOML
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// EnsureDirectories creates all necessary directories
func EnsureDirectories() error {
	dirs := []func() (string, error){
		DataPath,
		CachePath,
		func() (string, error) {
			configPath, err := ConfigPath()
			if err != nil {
				return "", err
			}
			return filepath.Dir(configPath), nil
		},
	}

	for _, dirFunc := range dirs {
		dir, err := dirFunc()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetScriptConfig returns the effective configuration for a specific script URL
func (c *ConfigWithOverrides) GetScriptConfig(url string) *Config {
	// Start with base config
	config := c.Config

	// Apply script-specific overrides if they exist
	if override, exists := c.Scripts[url]; exists {
		// Simple merge - override non-zero values
		if override.Cache.Strategy != "" {
			config.Cache.Strategy = override.Cache.Strategy
		}
		if override.Cache.TTL != 0 {
			config.Cache.TTL = override.Cache.TTL
		}
		if override.Security.VerifyHTTPS != config.Security.VerifyHTTPS {
			config.Security.VerifyHTTPS = override.Security.VerifyHTTPS
		}
		if override.Security.PromptOnChange != config.Security.PromptOnChange {
			config.Security.PromptOnChange = override.Security.PromptOnChange
		}
		if override.Rollback.KeepVersions != 0 {
			config.Rollback.KeepVersions = override.Rollback.KeepVersions
		}
	}

	return &config
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	validStrategies := map[string]bool{
		"smart":   true,
		"always":  true,
		"pin":     true,
		"offline": true,
	}

	if !validStrategies[c.Cache.Strategy] {
		return fmt.Errorf("invalid cache strategy: %s", c.Cache.Strategy)
	}

	if c.Cache.TTL < 0 {
		return fmt.Errorf("cache TTL must be positive")
	}

	if c.Rollback.KeepVersions < 0 {
		return fmt.Errorf("keep_versions must be positive")
	}

	if c.Network.Timeout < 0 {
		return fmt.Errorf("network timeout must be positive")
	}

	if c.Network.Retries < 0 {
		return fmt.Errorf("network retries must be positive")
	}

	return nil
}
