# CLI Reference

Complete command-line reference for the Elastic Integration Corpus Generator Tool.

## Global Usage

```bash
elastic-integration-corpus-generator-tool [command] [flags]
```

If running from source:
```bash
go run main.go [command] [flags]
```

## Commands Overview

| Command | Description |
|---------|-------------|
| `generate` | Generate corpus from Elastic Package Registry |
| `generate-with-template` | Generate corpus from custom template files |
| `local-template` | Generate corpus from built-in templates |
| `version` | Show application version |
| `completion` | Generate shell autocompletion scripts |
| `help` | Help about any command |

---

## `generate`

Generate a corpus for a specific integration data stream from the Elastic Package Registry.

### Usage

```bash
elastic-integration-corpus-generator-tool generate <package> <dataset> <version> [flags]
```

### Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `package` | Integration package name (e.g., `aws`, `kubernetes`) | Yes |
| `dataset` | Data stream name (e.g., `dynamodb`, `pod`) | Yes |
| `version` | Package version (e.g., `1.28.3`) | Yes |

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--tot-events` | `-t` | uint64 | `1` | Total events to generate (0 = infinite) |
| `--config-file` | `-c` | string | `""` | Path to field configuration file |
| `--package-registry-base-url` | `-r` | string | `https://epr.elastic.co/` | Package registry URL |
| `--now` | `-n` | string | `""` | Base timestamp for date fields |
| `--seed` | `-s` | int64 | `1` | Random seed for reproducibility |

### Examples

```bash
# Generate 1000 AWS DynamoDB events
elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 1000

# Generate with custom configuration
elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 1000 -c ./config.yml

# Generate with specific timestamp
elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 100 -n "2024-01-15T10:00:00.000000Z"

# Generate reproducible data with seed
elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 100 -s 42

# Generate infinite events (Ctrl+C to stop)
elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 0
```

---

## `generate-with-template`

Generate a corpus using custom template and field definition files.

### Usage

```bash
elastic-integration-corpus-generator-tool generate-with-template <template-path> <fields-path> [flags]
```

### Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `template-path` | Path to template file (`.tpl`) | Yes |
| `fields-path` | Path to fields definition file (`.yml`) | Yes |

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--tot-events` | `-t` | uint64 | `1` | Total events to generate (0 = infinite) |
| `--config-file` | `-c` | string | `""` | Path to field configuration file |
| `--template-type` | `-y` | string | `placeholder` | Template engine: `placeholder` or `gotext` |
| `--now` | `-n` | string | `""` | Base timestamp for date fields |
| `--seed` | `-s` | int64 | `1` | Random seed for reproducibility |

### Examples

```bash
# Generate using gotext template
elastic-integration-corpus-generator-tool generate-with-template \
  ./assets/templates/aws.vpcflow/schema-a/gotext.tpl \
  ./assets/templates/aws.vpcflow/schema-a/fields.yml \
  -t 1000 \
  -y gotext

# Generate with configuration
elastic-integration-corpus-generator-tool generate-with-template \
  ./my-template.tpl \
  ./my-fields.yml \
  -t 1000 \
  -y gotext \
  -c ./my-config.yml

# Use placeholder engine for maximum performance
elastic-integration-corpus-generator-tool generate-with-template \
  ./template.tpl \
  ./fields.yml \
  -t 1000000 \
  -y placeholder
```

---

## `local-template`

Generate a corpus from built-in templates in the `assets/templates/` folder.

### Usage

```bash
elastic-integration-corpus-generator-tool local-template <package> <dataset> [flags]
```

### Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `package` | Integration package name (e.g., `aws`, `kubernetes`) | Yes |
| `dataset` | Data stream name (e.g., `vpcflow`, `pod`) | Yes |

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--tot-events` | `-t` | uint64 | `1` | Total events to generate (0 = infinite) |
| `--config-file` | `-c` | string | `""` | Path to field configuration file |
| `--engine` | `-e` | string | `gotext` | Template engine: `placeholder` or `gotext` |
| `--schema` | | string | `b` | Data schema: `a` or `b` |
| `--now` | `-n` | string | `""` | Base timestamp for date fields |
| `--seed` | `-s` | int64 | `1` | Random seed for reproducibility |

### Available Templates

| Package | Dataset | Schema A | Schema B | Status |
|---------|---------|----------|----------|--------|
| `aws` | `billing` | ❌ | ✅ | Working |
| `aws` | `ec2_logs` | ❌ | ✅ | Working |
| `aws` | `ec2_metrics` | ❌ | ✅ | Has issues |
| `aws` | `sqs` | ❌ | ✅ | Working |
| `aws` | `vpcflow` | ✅ | ❌ | Working |
| `kubernetes` | `container` | ❌ | ✅ | Has issues |
| `kubernetes` | `pod` | ❌ | ✅ | Has issues |

> **Note:** For templates with issues, use the `generate` command with the package registry instead.

### Examples

```bash
# Generate AWS VPC Flow Logs (Schema A - raw format)
elastic-integration-corpus-generator-tool local-template aws vpcflow -t 1000 --schema a

# Generate Kubernetes Pod metrics (Schema B - JSON)
elastic-integration-corpus-generator-tool local-template kubernetes pod -t 1000 --schema b

# Generate with placeholder engine for speed
elastic-integration-corpus-generator-tool local-template aws vpcflow -t 100000 --schema a -e placeholder

# Generate with custom configuration
elastic-integration-corpus-generator-tool local-template aws ec2_metrics -t 1000 --schema b -c ./custom-config.yml
```

---

## `version`

Display the application version and build information.

### Usage

```bash
elastic-integration-corpus-generator-tool version
```

### Output Example

```
elastic-integration-corpus-generator-tool v1.0.0 version-hash abc123-dirty (source date: 2024-01-15T10:00:00Z)
```

---

## `help`

Get help for any command.

### Usage

```bash
# General help
elastic-integration-corpus-generator-tool --help
elastic-integration-corpus-generator-tool help

# Command-specific help
elastic-integration-corpus-generator-tool generate --help
elastic-integration-corpus-generator-tool help generate
```

---

## `completion`

Generate shell autocompletion scripts.

### Usage

```bash
# For Bash
elastic-integration-corpus-generator-tool completion bash > /etc/bash_completion.d/elastic-integration-corpus-generator-tool

# For Zsh
elastic-integration-corpus-generator-tool completion zsh > "${fpath[1]}/_elastic-integration-corpus-generator-tool"

# For Fish
elastic-integration-corpus-generator-tool completion fish > ~/.config/fish/completions/elastic-integration-corpus-generator-tool.fish

# For PowerShell
elastic-integration-corpus-generator-tool completion powershell > elastic-integration-corpus-generator-tool.ps1
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ELASTIC_INTEGRATION_CORPUS` | Custom output directory for generated files |

### Example

```bash
export ELASTIC_INTEGRATION_CORPUS=/tmp/my-corpus
elastic-integration-corpus-generator-tool local-template aws vpcflow -t 100 --schema a
# Output: /tmp/my-corpus/1649330390-gotext.tpl
```

---

## Output

Generated files are saved as NDJSON (newline-delimited JSON) to:

- **macOS:** `~/Library/Application Support/elastic-integration-corpus-generator-tool/corpora/`
- **Linux:** `~/.local/share/elastic-integration-corpus-generator-tool/corpora/`

### Filename Format

- **From generate:** `<timestamp>-<package>-<dataset>-<version>.ndjson`
- **From template:** `<timestamp>-<template-filename>`

---

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | Error (invalid arguments, file not found, etc.) |

---

## Tips

### Generate Large Datasets

For generating large datasets efficiently:

```bash
# Use placeholder engine (3-9x faster)
elastic-integration-corpus-generator-tool generate-with-template ./template.tpl ./fields.yml -t 10000000 -y placeholder
```

### Reproducible Data

Use the `--seed` flag for reproducible output:

```bash
# These will produce identical output
elastic-integration-corpus-generator-tool local-template aws vpcflow -t 100 --schema a -s 42
elastic-integration-corpus-generator-tool local-template aws vpcflow -t 100 --schema a -s 42
```

### Infinite Generation

Generate events continuously (useful for streaming scenarios):

```bash
# Press Ctrl+C to stop
elastic-integration-corpus-generator-tool local-template kubernetes pod -t 0 --schema b
```

### Custom Timestamps

Set a specific base time for date fields:

```bash
elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 100 -n "2024-06-15T12:00:00.000000Z"
```

---

## See Also

- [Quick Start Guide](./quickstart.md)
- [Usage Guide](./usage.md)
- [Writing Templates](./writing-templates.md)
- [Fields Configuration](./fields-configuration.md)
