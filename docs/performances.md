# Performance Guide

This guide covers performance characteristics, benchmarks, and optimization strategies for the corpus generator.

## Table of Contents

- [Overview](#overview)
- [Template Engine Comparison](#template-engine-comparison)
- [Benchmarks](#benchmarks)
- [Optimization Tips](#optimization-tips)
- [Running Benchmarks](#running-benchmarks)

## Overview

Performance is a key consideration when generating large datasets. The tool offers two template engines optimized for different use cases:

| Engine | Speed | Features | Best For |
|--------|-------|----------|----------|
| `placeholder` | Fastest | Basic substitution | Large datasets, production |
| `gotext` | 3-9x slower | Full template logic | Development, complex templates |

## Template Engine Comparison

### Placeholder Engine

**Pros:**
- Maximum throughput
- Minimal memory allocation
- Simple and predictable

**Cons:**
- No conditional logic
- No calculations
- No helper functions
- Limited flexibility

**Use when:**
- Generating millions of events
- Performance is critical
- Template is simple

### GoText Engine

**Pros:**
- Full Go template syntax
- Sprig helper functions
- Conditional logic
- Variable storage
- Date manipulation

**Cons:**
- 3-9x slower than placeholder
- Higher memory usage
- More allocations

**Use when:**
- Template requires logic
- Development and testing
- Complex data relationships

## Benchmarks

Benchmarks from PR #40, comparing three generator implementations:

### Test Scenarios

1. **JSONContent**: Schema C data for "endpoint process 8.2.0" integration
2. **VPCFlowLogs**: Schema A data for AWS VPC flow logs

### Results (16-core machine)

#### Time per Operation

| Generator | JSONContent | VPCFlowLogs |
|-----------|-------------|-------------|
| Legacy | 47.7 microseconds | - |
| Placeholder | 30.0 microseconds | 1.09 microseconds |
| GoText | 281 microseconds | 12.8 microseconds |

**Key insight:** Placeholder is ~9x faster than GoText for JSON content, ~12x faster for simple log lines.

#### Memory Allocation per Operation

| Generator | JSONContent | VPCFlowLogs |
|-----------|-------------|-------------|
| Legacy | 3.82 KB | - |
| Placeholder | 432 bytes | 64 bytes |
| GoText | 48.3 KB | 2.32 KB |

**Key insight:** Placeholder uses ~100x less memory than GoText.

#### Allocations per Operation

| Generator | JSONContent | VPCFlowLogs |
|-----------|-------------|-------------|
| Legacy | 22 | - |
| Placeholder | 14 | 2 |
| GoText | 2,230 | 95 |

### Real-World Test: 20GB Dataset

Generating 20GB of "aws dynamodb 1.28.3" Schema C data:

| Generator | Time |
|-----------|------|
| Legacy | 1m 44s |
| Placeholder | 1m 34s |
| GoText | 6m 50s |

**Key insight:** For large datasets, placeholder is ~4x faster than GoText.

## Optimization Tips

### 1. Choose the Right Engine

```bash
# For large datasets, use placeholder
./elastic-integration-corpus-generator-tool generate-with-template \
  template.tpl fields.yml -t 10000000 -y placeholder

# For development, use gotext
./elastic-integration-corpus-generator-tool generate-with-template \
  template.tpl fields.yml -t 1000 -y gotext
```

### 2. Simplify Templates

Reduce template complexity where possible:

```text
# Slower: Multiple function calls
{{- $ts := generate "timestamp" -}}
{{- $formatted := $ts.Format "2006-01-02T15:04:05Z07:00" -}}
{"timestamp": "{{$formatted}}"}

# Faster: Inline formatting
{"timestamp": "{{(generate "timestamp").Format "2006-01-02T15:04:05Z07:00"}}"}
```

### 3. Minimize Variables

Each variable allocation has overhead:

```text
# Slower: Many variables
{{- $a := generate "field_a" -}}
{{- $b := generate "field_b" -}}
{{- $c := generate "field_c" -}}
{"a": "{{$a}}", "b": "{{$b}}", "c": "{{$c}}"}

# Faster: Inline generation
{"a": "{{generate "field_a"}}", "b": "{{generate "field_b"}}", "c": "{{generate "field_c"}}"}
```

### 4. Reduce Cardinality Lookups

High cardinality with many fields increases lookup time:

```yaml
# Consider if you really need high cardinality
fields:
  - name: request.id
    cardinality: 1000000  # High overhead

  - name: request.id
    cardinality: 10000    # Lower overhead, still realistic
```

### 5. Use Appropriate Event Counts

Generate only what you need:

```bash
# Testing: Small dataset
./elastic-integration-corpus-generator-tool local-template aws vpcflow -t 100 --schema a

# Benchmarking: Large dataset
./elastic-integration-corpus-generator-tool local-template aws vpcflow -t 10000000 --schema a
```

### 6. Consider Disk I/O

For very large datasets, disk write speed may become the bottleneck:

```bash
# Write to fast storage (SSD, tmpfs)
export ELASTIC_INTEGRATION_CORPUS=/tmp/corpus
./elastic-integration-corpus-generator-tool local-template aws ec2_metrics -t 10000000 --schema b
```

## Running Benchmarks

### Built-in Benchmarks

Run the benchmark test suite:

```bash
cd pkg/genlib
go test -bench=. -benchmem
```

### Custom Benchmarks

Time your specific use case:

```bash
time ./elastic-integration-corpus-generator-tool generate-with-template \
  ./assets/templates/aws.vpcflow/schema-a/gotext.tpl \
  ./assets/templates/aws.vpcflow/schema-a/fields.yml \
  -t 1000000 \
  -y gotext
```

### Comparing Engines

```bash
# Placeholder
time ./elastic-integration-corpus-generator-tool generate-with-template \
  template.tpl fields.yml -t 1000000 -y placeholder

# GoText
time ./elastic-integration-corpus-generator-tool generate-with-template \
  template.tpl fields.yml -t 1000000 -y gotext
```

## Performance Expectations

| Events | Placeholder | GoText |
|--------|-------------|--------|
| 1,000 | < 1 second | < 1 second |
| 100,000 | ~1 second | ~5 seconds |
| 1,000,000 | ~10 seconds | ~50 seconds |
| 10,000,000 | ~2 minutes | ~8 minutes |

*Actual times vary based on template complexity and hardware.*

## See Also

- [Writing Templates](./writing-templates.md) - Template creation guide
- [CLI Reference](./cli-help.md) - Command options
- [Usage Guide](./usage.md) - Common use cases
