# Tira 👑
> Universal package manager for the curl | bash era

Tira is a universal package manager that brings version control and
rollback capabilities to curl | bash installation scripts. Install
Ollama, Docker, NVM, Nix, or any script-based tool with confidence,
knowing you can always roll back if something breaks.

## Quick Start

Install Tira:
```bash
curl -fsSL https://tira.sh | sh
```

Use Tira to manage any installation:
```bash
# Instead of: curl -fsSL https://ollama.com/install.sh | sh
tira install https://ollama.com/install.sh

# List what's installed
tira list

# Rollback if needed
tira rollback ollama

# Upgrade everything
tira upgrade
```

## Why Tira?

- **📦 Version tracking** - Know exactly what version of each script you ran
- **🔄 Safe rollbacks** - Go back to previous versions if something breaks
- **🔍 Filesystem monitoring** - Track what files each script touched
- **⚡ Zero friction** - Works with any existing curl | bash script
- **🛡️ Script verification** - Detect when upstream scripts change

## Features

### Install & Track
```bash
tira install https://get.docker.com
# Downloads, caches, and runs the script
# Tracks what files were created/modified
```

### List Installed Packages
```bash
tira list
# ollama v1.2.0 (script hash: a1b2c3d4...)
# docker v24.0.7 (script hash: e5f6g7h8...)
```

### Rollback to Previous Versions
```bash
tira rollback ollama
# Removes current version, restores previous
```

### Update Detection
```bash
tira upgrade
# Checks all scripts for updates
# Prompts before installing new versions
```

## Configuration

Configure Tira behavior in `~/.config/tira/config.toml`:

```toml
[cache]
strategy = "smart"  # "smart" | "always" | "pin"
ttl = "1w"

[security]
verify_https = true
prompt_on_change = true

[rollback]
keep_versions = 5
```

## Installation

### Quick Install (Recommended)
```bash
curl -fsSL https://tira.sh | sh
```

### Manual Install
1. Download the latest release from [GitHub Releases](https://github.com/yourusername/tira/releases)
2. Extract and move to your PATH:
   ```bash
   tar -xzf tira-linux-amd64.tar.gz
   sudo mv tira /usr/local/bin/
   ```

## Examples

### GPU-optimized tools (perfect for modern hardware)
```bash
tira install https://ollama.com/install.sh      # AI models
tira install https://get.docker.com             # Containers
```

### Development tools
```bash
tira install https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh   # Node Version Manager
tira install https://nixos.org/nix/install                                      # Nix Package Manager
tira install https://sh.rustup.rs                                               # Rust
tira install https://deno.land/install.sh                                       # Deno
```

### Pin to specific versions
```bash
tira install https://ollama.com/install.sh@a1b2c3d4
```

## How it Works

1. **Intercepts** - When you run `tira install <url>`, Tira downloads the script
2. **Analyzes** - Computes hash, checks for changes from cached versions
3. **Monitors** - Tracks filesystem changes during installation
4. **Caches** - Stores script and installation metadata for rollbacks
5. **Manages** - Provides version control for your installations

## Contributing

We love contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

---

*Rule your installs* 👑
