/*
Copyright © 2025 Erik Wright <yo@oheriko.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/oheriko/tira/internal/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Tira configuration",
	Long: `View and modify Tira configuration settings.

Examples:
  tira config show                        # Show current configuration
  tira config set cache.strategy pin      # Set cache strategy
  tira config set cache.ttl 1w           # Set cache TTL to 1 week
  tira config set security.verify_https true
  tira config edit                        # Edit config file in $EDITOR`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current Tira configuration with all settings.`,
	Run:   configShowHandler,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long: `Set a configuration value using dot notation.

Examples:
  tira config set cache.strategy pin
  tira config set cache.ttl 24h
  tira config set cache.auto_cleanup true
  tira config set security.verify_https false
  tira config set security.prompt_on_change true
  tira config set rollback.keep_versions 10
  tira config set ui.show_hashes true
  tira config set ui.use_color false
  tira config set network.timeout 60s
  tira config set network.retries 5`,
	Args: cobra.ExactArgs(2),
	Run:  configSetHandler,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long: `Get a specific configuration value.

Examples:
  tira config get cache.strategy
  tira config get security.verify_https`,
	Args: cobra.ExactArgs(1),
	Run:  configGetHandler,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file in default editor",
	Long:  `Open the configuration file in your default editor ($EDITOR).`,
	Run:   configEditHandler,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Long:  `Display the path to the configuration file.`,
	Run:   configPathHandler,
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Add config subcommands
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configPathCmd)
}

func configShowHandler(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("📋 Current Tira configuration:")

	fmt.Printf("\n[cache]\n")
	fmt.Printf("strategy = %q\n", cfg.Cache.Strategy)
	fmt.Printf("ttl = %q\n", cfg.Cache.TTL.String())
	fmt.Printf("auto_cleanup = %t\n", cfg.Cache.AutoCleanup)

	fmt.Printf("\n[security]\n")
	fmt.Printf("verify_https = %t\n", cfg.Security.VerifyHTTPS)
	fmt.Printf("prompt_on_change = %t\n", cfg.Security.PromptOnChange)
	fmt.Printf("max_script_size = %d\n", cfg.Security.MaxScriptSize)
	fmt.Printf("require_checksum = %t\n", cfg.Security.RequireChecksum)
	if len(cfg.Security.BlockedDomains) > 0 {
		fmt.Printf("blocked_domains = %q\n", cfg.Security.BlockedDomains)
	}
	if len(cfg.Security.AllowedDomains) > 0 {
		fmt.Printf("allowed_domains = %q\n", cfg.Security.AllowedDomains)
	}

	fmt.Printf("\n[rollback]\n")
	fmt.Printf("keep_versions = %d\n", cfg.Rollback.KeepVersions)
	fmt.Printf("auto_backup = %t\n", cfg.Rollback.AutoBackup)

	fmt.Printf("\n[ui]\n")
	fmt.Printf("show_progress = %t\n", cfg.UI.ShowProgress)
	fmt.Printf("show_hashes = %t\n", cfg.UI.ShowHashes)
	fmt.Printf("confirm_installs = %t\n", cfg.UI.ConfirmInstalls)
	fmt.Printf("use_color = %t\n", cfg.UI.UseColor)
	fmt.Printf("verbose = %t\n", cfg.UI.Verbose)

	fmt.Printf("\n[network]\n")
	fmt.Printf("timeout = %q\n", cfg.Network.Timeout.String())
	fmt.Printf("retries = %d\n", cfg.Network.Retries)
	fmt.Printf("user_agent = %q\n", cfg.Network.UserAgent)
	if cfg.Network.Proxy != "" {
		fmt.Printf("proxy = %q\n", cfg.Network.Proxy)
	}
	fmt.Printf("skip_verify = %t\n", cfg.Network.SkipVerify)

	// Show script overrides if any exist
	if len(cfg.Scripts) > 0 {
		fmt.Printf("\n[script overrides]\n")
		for url := range cfg.Scripts {
			fmt.Printf("  %s\n", url)
		}
	}

	configPath, _ := config.ConfigPath()
	fmt.Printf("\n📁 Config file: %s\n", configPath)
}

func configSetHandler(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Parse key using dot notation
	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		fmt.Fprintf(os.Stderr, "❌ Error: key must be in format 'section.key'\n")
		fmt.Fprintf(os.Stderr, "   Example: cache.strategy, ui.show_hashes\n")
		os.Exit(1)
	}

	section, keyName := parts[0], parts[1]

	// Set the value based on section and key
	if err := setConfigValue(cfg, section, keyName, value); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}

	// Validate config
	if err := cfg.Config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: invalid config value: %v\n", err)
		os.Exit(1)
	}

	// Save config
	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Set %s = %s\n", key, value)
}

func configGetHandler(cmd *cobra.Command, args []string) {
	key := args[0]

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error loading config: %v\n", err)
		os.Exit(1)
	}

	value, err := getConfigValue(cfg, key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(value)
}

func configEditHandler(cmd *cobra.Command, args []string) {
	configPath, err := config.ConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error getting config path: %v\n", err)
		os.Exit(1)
	}

	// Create config file if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := &config.ConfigWithOverrides{
			Config:  *config.DefaultConfig(),
			Scripts: make(map[string]config.ScriptOverride),
		}
		if err := cfg.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error creating config file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Created new config file\n")
	}

	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Try common editors
		editors := []string{"nano", "vim", "vi", "emacs", "code"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}

	if editor == "" {
		fmt.Fprintf(os.Stderr, "❌ No editor found. Set $EDITOR environment variable.\n")
		fmt.Fprintf(os.Stderr, "   Example: export EDITOR=nano\n")
		fmt.Printf("📁 Config file location: %s\n", configPath)
		os.Exit(1)
	}

	fmt.Printf("📝 Opening config file in %s...\n", editor)

	// Execute editor
	editorCmd := exec.Command(editor, configPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error running editor: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Config file updated\n")
}

func configPathHandler(cmd *cobra.Command, args []string) {
	configPath, err := config.ConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error getting config path: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(configPath)
}

func setConfigValue(cfg *config.ConfigWithOverrides, section, key, value string) error {
	switch section {
	case "cache":
		switch key {
		case "strategy":
			cfg.Cache.Strategy = value
		case "ttl":
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration format: %s (use format like '24h', '1w')", value)
			}
			cfg.Cache.TTL = duration
		case "auto_cleanup":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value: %s (use 'true' or 'false')", value)
			}
			cfg.Cache.AutoCleanup = val
		default:
			return fmt.Errorf("unknown cache key: %s", key)
		}
	case "security":
		switch key {
		case "verify_https":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value: %s", value)
			}
			cfg.Security.VerifyHTTPS = val
		case "prompt_on_change":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value: %s", value)
			}
			cfg.Security.PromptOnChange = val
		case "max_script_size":
			val, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value: %s", value)
			}
			cfg.Security.MaxScriptSize = val
		case "require_checksum":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value: %s", value)
			}
			cfg.Security.RequireChecksum = val
		default:
			return fmt.Errorf("unknown security key: %s", key)
		}
	case "rollback":
		switch key {
		case "keep_versions":
			val, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid integer value: %s", value)
			}
			cfg.Rollback.KeepVersions = val
		case "auto_backup":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value: %s", value)
			}
			cfg.Rollback.AutoBackup = val
		default:
			return fmt.Errorf("unknown rollback key: %s", key)
		}
	case "ui":
		switch key {
		case "show_progress":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value: %s", value)
			}
			cfg.UI.ShowProgress = val
		case "show_hashes":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value: %s", value)
			}
			cfg.UI.ShowHashes = val
		case "confirm_installs":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value: %s", value)
			}
			cfg.UI.ConfirmInstalls = val
		case "use_color":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value: %s", value)
			}
			cfg.UI.UseColor = val
		case "verbose":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value: %s", value)
			}
			cfg.UI.Verbose = val
		default:
			return fmt.Errorf("unknown ui key: %s", key)
		}
	case "network":
		switch key {
		case "timeout":
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration format: %s", value)
			}
			cfg.Network.Timeout = duration
		case "retries":
			val, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid integer value: %s", value)
			}
			cfg.Network.Retries = val
		case "user_agent":
			cfg.Network.UserAgent = value
		case "proxy":
			cfg.Network.Proxy = value
		case "skip_verify":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid boolean value: %s", value)
			}
			cfg.Network.SkipVerify = val
		default:
			return fmt.Errorf("unknown network key: %s", key)
		}
	default:
		return fmt.Errorf("unknown section: %s", section)
	}

	return nil
}

func getConfigValue(cfg *config.ConfigWithOverrides, key string) (string, error) {
	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("key must be in format 'section.key'")
	}

	section, keyName := parts[0], parts[1]

	switch section {
	case "cache":
		switch keyName {
		case "strategy":
			return cfg.Cache.Strategy, nil
		case "ttl":
			return cfg.Cache.TTL.String(), nil
		case "auto_cleanup":
			return strconv.FormatBool(cfg.Cache.AutoCleanup), nil
		default:
			return "", fmt.Errorf("unknown cache key: %s", keyName)
		}
	case "security":
		switch keyName {
		case "verify_https":
			return strconv.FormatBool(cfg.Security.VerifyHTTPS), nil
		case "prompt_on_change":
			return strconv.FormatBool(cfg.Security.PromptOnChange), nil
		case "max_script_size":
			return strconv.FormatInt(cfg.Security.MaxScriptSize, 10), nil
		case "require_checksum":
			return strconv.FormatBool(cfg.Security.RequireChecksum), nil
		default:
			return "", fmt.Errorf("unknown security key: %s", keyName)
		}
	case "rollback":
		switch keyName {
		case "keep_versions":
			return strconv.Itoa(cfg.Rollback.KeepVersions), nil
		case "auto_backup":
			return strconv.FormatBool(cfg.Rollback.AutoBackup), nil
		default:
			return "", fmt.Errorf("unknown rollback key: %s", keyName)
		}
	case "ui":
		switch keyName {
		case "show_progress":
			return strconv.FormatBool(cfg.UI.ShowProgress), nil
		case "show_hashes":
			return strconv.FormatBool(cfg.UI.ShowHashes), nil
		case "confirm_installs":
			return strconv.FormatBool(cfg.UI.ConfirmInstalls), nil
		case "use_color":
			return strconv.FormatBool(cfg.UI.UseColor), nil
		case "verbose":
			return strconv.FormatBool(cfg.UI.Verbose), nil
		default:
			return "", fmt.Errorf("unknown ui key: %s", keyName)
		}
	case "network":
		switch keyName {
		case "timeout":
			return cfg.Network.Timeout.String(), nil
		case "retries":
			return strconv.Itoa(cfg.Network.Retries), nil
		case "user_agent":
			return cfg.Network.UserAgent, nil
		case "proxy":
			return cfg.Network.Proxy, nil
		case "skip_verify":
			return strconv.FormatBool(cfg.Network.SkipVerify), nil
		default:
			return "", fmt.Errorf("unknown network key: %s", keyName)
		}
	default:
		return "", fmt.Errorf("unknown section: %s", section)
	}
}
