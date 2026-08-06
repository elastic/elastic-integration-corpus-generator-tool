# Contributing Guide

Thank you for your interest in contributing to the Elastic Integration Corpus Generator Tool!

## Prerequisites

- **Go 1.18+** - [Download Go](https://go.dev/dl/)
- **Git** - For version control
- **Make** - For build automation

## Development Setup

```bash
# Clone the repository
git clone https://github.com/elastic/elastic-integration-corpus-generator-tool.git
cd elastic-integration-corpus-generator-tool

# Install dependencies
go mod download

# Build
make build

# Verify
./elastic-integration-corpus-generator-tool version
```

## Building

```bash
# Standard build with version info
make build

# Quick build without version info
go build -o elastic-integration-corpus-generator-tool

# Run without building
go run main.go --help
```

## Testing

```bash
# Run all tests
make test

# Run specific tests
go test -v ./pkg/genlib/...

# Run benchmarks
go test -bench=. -benchmem ./pkg/genlib/
```

## Code Style

```bash
# Format code
go fmt ./...

# Check for issues
go vet ./...

# Add license headers
make licenser
```

All source files (except `assets/`) must have the Elastic License 2.0 header.

## Adding Templates

Create templates in `assets/templates/<package>.<dataset>/schema-<x>/`:

```
assets/templates/mypackage.mydataset/schema-b/
  fields.yml      # Required: Field definitions
  configs.yml     # Optional: Generation config
  gotext.tpl      # Required: GoText template
```

Test your template:

```bash
go run main.go local-template mypackage mydataset -t 10 --schema b
```

## Submitting Changes

Before submitting:

1. Run `make test`
2. Run `make licenser`
3. Run `go fmt ./...`
4. Run `go vet ./...`

Then:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Open a Pull Request

## Release Process

See [Issue #77](https://github.com/elastic/elastic-integration-corpus-generator-tool/issues/77).

## Getting Help

- **Questions:** Open a [Discussion](https://github.com/elastic/elastic-integration-corpus-generator-tool/discussions)
- **Bugs:** Open an [Issue](https://github.com/elastic/elastic-integration-corpus-generator-tool/issues)

## License

Contributions are licensed under the [Elastic License 2.0](./LICENSE.md).
