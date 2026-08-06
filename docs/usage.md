# Usage Guide

This guide covers common use cases and practical examples for the Elastic Integration Corpus Generator Tool.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Use Case 1: Testing Ingest Pipelines](#use-case-1-testing-ingest-pipelines)
- [Use Case 2: Performance Benchmarking](#use-case-2-performance-benchmarking)
- [Use Case 3: Integration Development](#use-case-3-integration-development)
- [Use Case 4: Generating Realistic Test Data](#use-case-4-generating-realistic-test-data)
- [Use Case 5: Creating Rally Tracks](#use-case-5-creating-rally-tracks)
- [Advanced Scenarios](#advanced-scenarios)

## Basic Usage

### Generate Schema C Data from Package Registry

The `generate` command downloads field definitions from the Elastic Package Registry and generates Schema C (post-ingest pipeline) data:

```bash
./elastic-integration-corpus-generator-tool generate <package> <dataset> <version> --tot-events <quantity>
```

**Arguments:**
- `package`: Integration package name (e.g., `aws`, `kubernetes`)
- `dataset`: Data stream name (e.g., `dynamodb`, `pod`)
- `version`: Package version (e.g., `1.28.3`)

**Example:**

```bash
./elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 1000 --config-file config.yml
# Output: File generated: /path/to/corpora/1649330390-aws-dynamodb-1.28.3.ndjson
```

### Generate Schema A/B Data from Templates

The `generate-with-template` command uses template files for more control over output format:

```bash
./elastic-integration-corpus-generator-tool generate-with-template <template-path> <fields-path> --tot-events <quantity>
```

**Example:**

```bash
./elastic-integration-corpus-generator-tool generate-with-template \
  ./assets/templates/aws.vpcflow/schema-a/gotext.tpl \
  ./assets/templates/aws.vpcflow/schema-a/fields.yml \
  -t 1000 \
  --config-file ./assets/templates/aws.vpcflow/schema-a/configs.yml \
  -y gotext
# Output: File generated: /path/to/corpora/1684304483-gotext.tpl
```

### Generate from Built-in Templates

The `local-template` command provides a shortcut for built-in templates:

```bash
./elastic-integration-corpus-generator-tool local-template aws vpcflow -t 1000 --schema a
```

## Use Case 1: Testing Ingest Pipelines

Generate Schema B data to test your Elasticsearch ingest pipelines without needing a running Elastic Agent.

### Scenario

You want to test an ingest pipeline for Kubernetes pod metrics.

### Steps

1. **Generate Schema B data:**

```bash
./elastic-integration-corpus-generator-tool local-template kubernetes pod -t 1000 --schema b
```

2. **Index the data using the Bulk API:**

```bash
curl -X POST "localhost:9200/test-index/_bulk?pipeline=kubernetes-pod-pipeline" \
  -H "Content-Type: application/x-ndjson" \
  --data-binary @~/.local/share/elastic-integration-corpus-generator-tool/corpora/*.ndjson
```

3. **Verify the pipeline processed the data correctly:**

```bash
curl "localhost:9200/test-index/_search?pretty"
```

## Use Case 2: Performance Benchmarking

Generate large datasets for performance testing Elasticsearch clusters.

### Scenario

You need to benchmark indexing performance with millions of AWS EC2 metrics events.

### Steps

1. **Generate a large dataset:**

```bash
# Use placeholder engine for maximum speed
./elastic-integration-corpus-generator-tool generate-with-template \
  ./assets/templates/aws.ec2_metrics/schema-b/gotext.tpl \
  ./assets/templates/aws.ec2_metrics/schema-b/fields.yml \
  -t 10000000 \
  -y placeholder \
  -c ./assets/templates/aws.ec2_metrics/schema-b/configs.yml
```

2. **Monitor generation performance:**

```bash
time ./elastic-integration-corpus-generator-tool local-template aws ec2_metrics -t 1000000 --schema b
```

### Performance Tips

- Use `placeholder` template engine for 3-9x faster generation
- See [Performance Guide](./performances.md) for detailed benchmarks

## Use Case 3: Integration Development

Generate test data while developing new Elastic integrations.

### Scenario

You are developing a new integration and need sample data for testing.

### Steps

1. **Create field definitions (`fields.yml`):**

```yaml
- name: timestamp
  type: date
- name: message
  type: keyword
- name: host.name
  type: keyword
- name: log.level
  type: keyword
- name: service.name
  type: constant_keyword
- name: http.response.status_code
  type: integer
- name: http.response.time
  type: double
```

2. **Create a template (`template.tpl`):**

```text
{"timestamp":"{{generate "timestamp"}}","message":"{{generate "message"}}","host":{"name":"{{generate "host.name"}}"},"log":{"level":"{{generate "log.level"}}"},"service":{"name":"{{generate "service.name"}}"},"http":{"response":{"status_code":{{generate "http.response.status_code"}},"time":{{generate "http.response.time"}}}}}
```

3. **Create configuration (`config.yml`):**

```yaml
fields:
  - name: timestamp
    period: "-1h"
  - name: log.level
    enum: ["INFO", "WARN", "ERROR", "DEBUG"]
  - name: service.name
    value: "my-service"
  - name: http.response.status_code
    enum: ["200", "201", "400", "404", "500"]
  - name: http.response.time
    range:
      min: 0.1
      max: 5.0
    fuzziness: 0.2
```

4. **Generate test data:**

```bash
./elastic-integration-corpus-generator-tool generate-with-template \
  ./template.tpl \
  ./fields.yml \
  -t 1000 \
  -y gotext \
  -c ./config.yml
```

## Use Case 4: Generating Realistic Test Data

Create data that mimics real-world patterns using cardinality and fuzziness settings.

### Scenario

Generate Kubernetes metrics that simulate a cluster with:
- 50 nodes
- 500 pods (10 pods per node)
- 20 namespaces

### Configuration

```yaml
# realistic-k8s-config.yml
fields:
  - name: kubernetes.node.name
    cardinality: 50
  - name: kubernetes.pod.name
    cardinality: 500
  - name: kubernetes.namespace
    cardinality: 20
  - name: kubernetes.pod.cpu.usage.nanocores
    range:
      min: 1000000
      max: 4000000000
    fuzziness: 0.1
  - name: kubernetes.pod.memory.usage.bytes
    range:
      min: 10485760
      max: 8589934592
    fuzziness: 0.15
```

### Generate

```bash
./elastic-integration-corpus-generator-tool local-template kubernetes pod -t 10000 --schema b -c realistic-k8s-config.yml
```

## Use Case 5: Creating Rally Tracks

Generate data for [Rally](https://github.com/elastic/rally) performance testing.

### Scenario

Create a Rally track with AWS billing data.

### Steps

1. **Generate the corpus:**

```bash
./elastic-integration-corpus-generator-tool local-template aws billing -t 100000 --schema b
```

2. **Create Rally track structure:**

```
my-track/
  track.json
  index.json
  documents/
    billing.json
```

3. **Use the generated file in your Rally track configuration.**

## Advanced Scenarios

### Reproducible Data Generation

Use the `--seed` flag for deterministic output:

```bash
# Both commands produce identical data
./elastic-integration-corpus-generator-tool local-template aws vpcflow -t 100 --schema a -s 12345
./elastic-integration-corpus-generator-tool local-template aws vpcflow -t 100 --schema a -s 12345
```

### Custom Timestamps

Set a specific base time for date fields:

```bash
./elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 100 \
  -n "2024-06-15T12:00:00.000000Z"
```

### Continuous Generation

Generate events indefinitely (useful for streaming scenarios):

```bash
# Press Ctrl+C to stop
./elastic-integration-corpus-generator-tool local-template kubernetes pod -t 0 --schema b
```

### Counter Fields with Reset

Generate ever-increasing values that periodically reset:

```yaml
fields:
  - name: network.bytes_sent
    counter: true
    fuzziness: 0.1
    counter_reset:
      strategy: "after_n"
      reset_after_n: 100
```

### Weighted Enum Distribution

Control probability distribution of enum values:

```yaml
fields:
  - name: http.response.status_code
    # 80% success (200), 10% client error (400), 10% server error (500)
    enum: ["200", "200", "200", "200", "200", "200", "200", "200", "400", "500"]
```

## Output Location

Generated files are saved to (varies by OS):
- **macOS:** `~/Library/Application Support/elastic-integration-corpus-generator-tool/corpora/`
- **Linux:** `~/.local/share/elastic-integration-corpus-generator-tool/corpora/`

Customize with environment variable:

```bash
export ELASTIC_INTEGRATION_CORPUS=/custom/path
```

## Next Steps

- [Writing Templates](./writing-templates.md) - Create custom templates
- [Fields Configuration](./fields-configuration.md) - Fine-tune data generation
- [CLI Reference](./cli-help.md) - Complete command reference
- [Performance Guide](./performances.md) - Optimize generation speed
