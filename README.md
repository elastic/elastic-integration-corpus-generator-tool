# Elastic Integration Corpus Generator Tool

<p align="center">
  <strong>Generate realistic synthetic event data for Elastic integrations</strong>
</p>

<p align="center">
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-features">Features</a> •
  <a href="#-installation">Installation</a> •
  <a href="#-usage">Usage</a> •
  <a href="#-documentation">Documentation</a> •
  <a href="#-contributing">Contributing</a>
</p>

---

A command-line tool for generating synthetic event corpus data for [Elastic integrations](https://www.elastic.co/integrations). Create realistic test data that mimics real observability data from AWS, Kubernetes, and other sources—perfect for testing ingest pipelines, benchmarking performance, and developing integrations.

## ⚡ Quick Start

```bash
# Build the tool
make build

# Generate 1000 events from a local template
./elastic-integration-corpus-generator-tool local-template aws vpcflow -t 1000 --schema a

# Generate events from the Elastic Package Registry
./elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 1000
```

**Output:** Generated data is saved to your OS data directory (e.g., `~/Library/Application Support/` on macOS, `~/.local/share/` on Linux)

## ✨ Features

- **🎯 Multiple Data Schemas** — Generate Schema A (raw logs), Schema B (agent-processed), or Schema C (post-ingest) data
- **🔧 Flexible Templates** — Use built-in templates or create custom ones for any integration
- **⚡ High Performance** — Two template engines: `placeholder` (fastest) and `gotext` (feature-rich)
- **🎲 Realistic Data** — Configure cardinality, fuzziness, ranges, and counters for life-like data
- **📦 Package Registry Integration** — Generate data directly from published Elastic integration packages
- **🔄 Reproducible** — Set random seeds for consistent, reproducible data generation

## 📦 Installation

### Prerequisites

- Go 1.18 or later
- `make` (for building)
- `git` (for version information)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/elastic/elastic-integration-corpus-generator-tool.git
cd elastic-integration-corpus-generator-tool

# Build the binary
make build

# Verify installation
./elastic-integration-corpus-generator-tool version
```

### Run without Building

```bash
go run main.go --help
```

## 🚀 Usage

### Generate from Local Templates

Use pre-built templates in the `assets/templates/` folder:

```bash
# Generate AWS VPC Flow Logs (Schema A - raw log format)
./elastic-integration-corpus-generator-tool local-template aws vpcflow -t 1000 --schema a

# Generate Kubernetes Pod metrics (Schema B - JSON format)
./elastic-integration-corpus-generator-tool local-template kubernetes pod -t 1000 --schema b
```

### Generate from Package Registry

Download field definitions from the Elastic Package Registry and generate Schema C data:

```bash
# Generate AWS DynamoDB metrics
./elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 1000

# With custom configuration
./elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 1000 -c config.yml
```

### Generate from Custom Templates

Use your own template and field definition files:

```bash
./elastic-integration-corpus-generator-tool generate-with-template \
  ./my-template.tpl \
  ./my-fields.yml \
  -t 1000 \
  --template-type gotext \
  --config-file ./my-config.yml
```

### Common Options

| Flag | Short | Description |
|------|-------|-------------|
| `--tot-events` | `-t` | Number of events to generate (0 = infinite) |
| `--config-file` | `-c` | Path to field configuration file |
| `--template-type` | `-y` | Template engine: `placeholder` or `gotext` |
| `--now` | `-n` | Base timestamp for date fields |
| `--seed` | `-s` | Random seed for reproducibility |
| `--schema` | | Data schema version: `a` or `b` |

## 📖 Documentation

| Document | Description |
|----------|-------------|
| [Quick Start Guide](./docs/quickstart.md) | Get started in 5 minutes |
| [Installation](./docs/installation.md) | Detailed installation instructions |
| [Usage Guide](./docs/usage.md) | Common use cases and examples |
| [CLI Reference](./docs/cli-help.md) | Complete command reference |
| [Writing Templates](./docs/writing-templates.md) | Create custom templates |
| [Field Types](./docs/field-types.md) | Supported Elasticsearch field types |
| [Fields Configuration](./docs/fields-configuration.md) | Configure data generation |
| [Template Helpers](./docs/go-text-template-helpers.md) | GoText template functions |
| [Data Schemas](./docs/data-schemas.md) | Understanding A/B/C/D schemas |
| [Cardinality](./docs/cardinality.md) | Control field value diversity |
| [Dimensionality](./docs/dimensionality.md) | Configure object fields |
| [Performance](./docs/performances.md) | Benchmarks and optimization |
| [Glossary](./docs/glossary.md) | Terminology reference |

## 🗂️ Available Templates

Built-in templates for common integrations:

| Integration | Dataset | Schemas | Status |
|-------------|---------|---------|--------|
| AWS | billing | B | ✅ |
| AWS | ec2_logs | B | ✅ |
| AWS | ec2_metrics | B | ⚠️ |
| AWS | sqs | B | ✅ |
| AWS | vpcflow | A | ✅ |
| Kubernetes | container | B | ⚠️ |
| Kubernetes | pod | B | ⚠️ |

> ⚠️ Some templates may have issues. Use `generate` command with package registry for most reliable results.

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

### Quick Development Commands

```bash
# Run tests
make test

# Add license headers
make licenser

# Build
make build
```

## 👥 Maintainers

[Observability Integrations Team](https://github.com/orgs/elastic/teams/obs-infraobs-integrations)

## 📄 License

[Elastic License 2.0](./LICENSE.md)
