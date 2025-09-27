# Tira
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

That's it! Tira installs itself and tracks its own installation. See what you have:

```bash
tira list
```
```
INSTALLED PACKAGES (1):
  ▓ tira installed-20250927-103630
    └─ Package installed from https://tira.sh/install.sh
    └─ https://tira.sh/install.sh
    └─ Installed 2 minutes ago (2025-09-27 10:36)
    └─ unknown
```

Now manage any curl | bash installation with full tracking:
```bash
# Install with confidence - Tira tracks everything
tira install https://ollama.com/install.sh
tira install https://get.docker.com
tira install https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh

# See your complete software inventory
tira list
```
```
INSTALLED PACKAGES (3):
  ▓ docker installed-20250927-104809
    └─ Container runtime platform
    └─ https://get.docker.com
    └─ Installed 5 minutes ago (2025-09-27 10:48)
    └─ containers, development, devops

  ▓ ollama installed-20250927-105307
    └─ Large language model runner
    └─ https://ollama.com/install.sh
    └─ Installed just now (2025-09-27 10:53)
    └─ ai, gpu, development, llm

  ▓ tira installed-20250927-103630
    └─ Package installed from https://tira.sh/install.sh
    └─ https://tira.sh/install.sh
    └─ Installed 16 minutes ago (2025-09-27 10:36)
    └─ unknown
```

## Why Tira?

- **Version tracking** - Know exactly what version of each script you ran
- **Safe rollbacks** - Go back to previous versions if something breaks
- **Filesystem monitoring** - Track what files each script touched
- **Zero friction** - Works with any existing curl | bash script
- **Script verification** - Detect when upstream scripts change

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
1. Download the latest release from [GitHub Releases](https://github.com/oheriko/tira/releases)
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

