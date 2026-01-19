# Quick Start Guide

Get up and running with the Elastic Integration Corpus Generator Tool in 5 minutes.

## Prerequisites

- Go 1.18 or later installed
- Terminal access

## Step 1: Get the Tool

```bash
# Clone the repository
git clone https://github.com/elastic/elastic-integration-corpus-generator-tool.git
cd elastic-integration-corpus-generator-tool

# Build the binary
make build
```

## Step 2: Generate Your First Data

Let's generate some AWS VPC Flow Log data:

```bash
./elastic-integration-corpus-generator-tool local-template aws vpcflow -t 100 --schema a
```

**Output:**
```
File generated: /home/user/.local/share/elastic-integration-corpus-generator-tool/corpora/1649330390-gotext.tpl
```

## Step 3: View the Generated Data

```bash
# View the first few lines
head -5 ~/.local/share/elastic-integration-corpus-generator-tool/corpora/*.ndjson
```

**Sample output (AWS VPC Flow Logs):**
```
2 627286350134 eni-abc123 192.168.1.100 10.0.0.50 54321 443 6 150 2250 2024-01-15T10:30:00.000000Z 2024-01-15T10:30:05.000000Z ACCEPT OK
2 627286350134 eni-def456 172.16.0.25 192.168.1.200 8080 80 6 42 630 2024-01-15T10:30:01.000000Z 2024-01-15T10:30:06.000000Z REJECT OK
```

## Step 4: Generate JSON Data

For Elasticsearch-ready JSON data, use Schema B:

```bash
./elastic-integration-corpus-generator-tool local-template kubernetes pod -t 10 --schema b
```

This generates NDJSON (newline-delimited JSON) ready for bulk indexing.

## Step 5: Customize Generation

Create a configuration file to control the generated data:

```yaml
# my-config.yml
fields:
  - name: SrcAddr
    cardinality: 50  # Only 50 unique source IPs
  - name: DstPort
    enum: ["80", "443", "8080"]  # Only these ports
```

Use it:
```bash
./elastic-integration-corpus-generator-tool local-template aws vpcflow -t 1000 --schema a -c my-config.yml
```

## What's Next?

- **[Usage Guide](./usage.md)** — More examples and use cases
- **[Writing Templates](./writing-templates.md)** — Create custom templates
- **[Fields Configuration](./fields-configuration.md)** — Fine-tune data generation
- **[CLI Reference](./cli-help.md)** — All commands and options

## Common Commands

```bash
# Generate from Elastic Package Registry
./elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 1000

# Generate with custom template
./elastic-integration-corpus-generator-tool generate-with-template ./template.tpl ./fields.yml -t 1000 -y gotext

# Show version
./elastic-integration-corpus-generator-tool version

# Get help
./elastic-integration-corpus-generator-tool --help
./elastic-integration-corpus-generator-tool generate --help
```

## Troubleshooting

### "dataset folder does not exist"

Make sure you're specifying a valid package and dataset. List available templates:
```bash
ls assets/templates/
```

### "template file does not exist"

Check that the schema folder contains the required template file:
```bash
ls assets/templates/aws.vpcflow/schema-a/
```

### Output location

Generated files are saved to (varies by OS):
- **macOS:** `~/Library/Application Support/elastic-integration-corpus-generator-tool/corpora/`
- **Linux:** `~/.local/share/elastic-integration-corpus-generator-tool/corpora/`
- **Custom:** Set `ELASTIC_INTEGRATION_CORPUS` environment variable
