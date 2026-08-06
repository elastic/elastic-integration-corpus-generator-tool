# Installation Guide

This guide covers all installation methods for the Elastic Integration Corpus Generator Tool.

## Prerequisites

### Required

- **Go 1.18+** — [Download Go](https://go.dev/dl/)
- **Git** — For cloning and version information
- **Make** — For building (included on most Unix systems)

### Verify Prerequisites

```bash
# Check Go version
go version
# Expected: go version go1.18 or higher

# Check Git
git --version

# Check Make
make --version
```

## Installation Methods

### Method 1: Build from Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/elastic/elastic-integration-corpus-generator-tool.git
cd elastic-integration-corpus-generator-tool

# Build the binary
make build

# Verify the build
./elastic-integration-corpus-generator-tool version
```

The binary will be created in the current directory as `elastic-integration-corpus-generator-tool`.

### Method 2: Go Install

```bash
go install github.com/elastic/elastic-integration-corpus-generator-tool@latest
```

> **Note:** This method won't include version information in the binary.

### Method 3: Run Without Building

For quick testing without building a binary:

```bash
git clone https://github.com/elastic/elastic-integration-corpus-generator-tool.git
cd elastic-integration-corpus-generator-tool

# Run directly with Go
go run main.go --help
go run main.go local-template aws vpcflow -t 100 --schema a
```

## Post-Installation Setup

### Add to PATH (Optional)

To run the tool from anywhere:

```bash
# Move to a directory in your PATH
sudo mv elastic-integration-corpus-generator-tool /usr/local/bin/

# Or add the current directory to PATH
export PATH=$PATH:$(pwd)
```

### Custom Output Location (Optional)

By default, generated files are saved to your OS data directory:
- **macOS:** `~/Library/Application Support/elastic-integration-corpus-generator-tool/corpora/`
- **Linux:** `~/.local/share/elastic-integration-corpus-generator-tool/corpora/`

To change this, set the environment variable:

```bash
# Temporary (current session)
export ELASTIC_INTEGRATION_CORPUS=/path/to/custom/location

# Permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export ELASTIC_INTEGRATION_CORPUS=/path/to/custom/location' >> ~/.bashrc
```

## Verify Installation

Run these commands to verify everything is working:

```bash
# Check version
./elastic-integration-corpus-generator-tool version

# Generate test data
./elastic-integration-corpus-generator-tool local-template aws vpcflow -t 10 --schema a

# Check output (path varies by OS)
# macOS:
ls -la ~/Library/Application\ Support/elastic-integration-corpus-generator-tool/corpora/
# Linux:
# ls -la ~/.local/share/elastic-integration-corpus-generator-tool/corpora/
```

## Directory Structure

After installation, the project structure looks like:

```
elastic-integration-corpus-generator-tool/
├── elastic-integration-corpus-generator-tool  # Built binary
├── main.go                                     # Entry point
├── cmd/                                        # CLI commands
├── pkg/genlib/                                 # Core library
├── internal/                                   # Internal packages
├── assets/templates/                           # Built-in templates
│   ├── aws.billing/
│   ├── aws.ec2_logs/
│   ├── aws.ec2_metrics/
│   ├── aws.sqs/
│   ├── aws.vpcflow/
│   ├── kubernetes.container/
│   └── kubernetes.pod/
└── docs/                                       # Documentation
```

## Updating

To update to the latest version:

```bash
cd elastic-integration-corpus-generator-tool
git pull origin main
make build
```

## Uninstalling

```bash
# Remove the binary
rm /usr/local/bin/elastic-integration-corpus-generator-tool
# Or from current directory
rm ./elastic-integration-corpus-generator-tool

# Remove generated data (optional)
rm -rf ~/.local/share/elastic-integration-corpus-generator-tool/

# Remove the source code (optional)
rm -rf /path/to/elastic-integration-corpus-generator-tool/
```

## Troubleshooting

### "go: command not found"

Go is not installed or not in your PATH. Install Go from [go.dev/dl](https://go.dev/dl/).

### "make: command not found"

On macOS:
```bash
xcode-select --install
```

On Ubuntu/Debian:
```bash
sudo apt-get install build-essential
```

On RHEL/CentOS:
```bash
sudo yum groupinstall "Development Tools"
```

### Build Errors

Ensure you have the correct Go version:
```bash
go version  # Should be 1.18 or higher
```

Try cleaning and rebuilding:
```bash
go clean -cache
make build
```

### Permission Denied

If you get permission errors when running the binary:
```bash
chmod +x elastic-integration-corpus-generator-tool
```

## Next Steps

- [Quick Start Guide](./quickstart.md) — Generate your first data
- [Usage Guide](./usage.md) — Common use cases
- [CLI Reference](./cli-help.md) — All commands and options
