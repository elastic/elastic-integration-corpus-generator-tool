# Fields Configuration

Fields configuration allows you to customize how data is generated for each field. This guide covers all available configuration options with examples.

## Table of Contents

- [Overview](#overview)
- [Configuration File Format](#configuration-file-format)
- [Configuration Options](#configuration-options)
  - [name](#name)
  - [value](#value)
  - [enum](#enum)
  - [range](#range)
  - [fuzziness](#fuzziness)
  - [cardinality](#cardinality)
  - [counter](#counter)
  - [counter_reset](#counter_reset)
  - [period](#period)
  - [object_keys](#object_keys)
- [Complete Example](#complete-example)
- [Common Patterns](#common-patterns)

## Overview

Configuration files (`configs.yml`) let you control:

- **What values** are generated (fixed values, enums, ranges)
- **How values change** over time (fuzziness, counters)
- **How many unique values** exist (cardinality)
- **When dates occur** (periods, date ranges)

## Configuration File Format

Configuration files are YAML with a root `fields` array:

```yaml
fields:
  - name: field_name
    # configuration options...
  - name: another_field
    # configuration options...
```

Place the file as `configs.yml` in your template folder, or specify a path with `--config-file`.

## Configuration Options

### `name`

**Required.** The dotted path to the field, matching an entry in `fields.yml`.

```yaml
fields:
  - name: host.name
  - name: aws.dynamodb.metrics.ReadCapacity.avg
  - name: data_stream.type
```

### `value`

Set a fixed value for the field. Overrides all other generation logic.

**Applies to:** All field types

```yaml
fields:
  # String value
  - name: data_stream.type
    value: metrics

  # Numeric value
  - name: version
    value: 2

  # Boolean value
  - name: enabled
    value: true

  # Array value (for indexed access in templates)
  - name: instance_types
    value: ["t2.micro", "t2.small", "t2.medium", "t2.large"]
```

**Template usage with array values:**

```yaml
# configs.yml
fields:
  - name: instance_type_index
    range:
      min: 0
      max: 3
  - name: instance_types
    value: ["t2.micro", "t2.small", "t2.medium", "t2.large"]
```

```text
{{- $idx := generate "instance_type_index" -}}
{{- $types := generate "instance_types" -}}
{"instance_type": "{{ index $types $idx }}"}
```

### `enum`

Randomly select from a list of allowed values.

**Applies to:** `keyword` type

```yaml
fields:
  # Equal probability
  - name: log.level
    enum: ["DEBUG", "INFO", "WARN", "ERROR"]

  # Weighted probability (repeat values)
  - name: http.response.status_code
    # 80% 200, 10% 400, 10% 500
    enum: ["200", "200", "200", "200", "200", "200", "200", "200", "400", "500"]

  # Combined with cardinality
  - name: action
    enum: ["ACCEPT", "REJECT"]
    cardinality: 2  # Only these 2 values
```

### `range`

Constrain generated values within bounds.

**Applies to:** `long`, `double`, `date` types

#### Numeric Range

```yaml
fields:
  # Integer range
  - name: http.response.status_code
    range:
      min: 100
      max: 599

  # Floating-point range
  - name: cpu.percent
    range:
      min: 0.0
      max: 100.0

  # One-sided range
  - name: bytes
    range:
      min: 0  # No upper bound
```

#### Date Range

```yaml
fields:
  # Specific date range
  - name: event.created
    range:
      from: "2024-01-01T00:00:00-00:00"
      to: "2024-01-31T23:59:59-00:00"

  # From a date to now
  - name: last_seen
    range:
      from: "2024-01-01T00:00:00-00:00"
      # 'to' defaults to time.Now()

  # From now to a future date
  - name: expires_at
    range:
      to: "2025-12-31T23:59:59-00:00"
      # 'from' defaults to time.Now()
```

**Date format:** `2006-01-02T15:04:05.999999999-07:00`

**Note:** Cannot combine `range` with `period` for date fields.

### `fuzziness`

Control how much consecutive values can change. Creates smoother, more realistic data progressions.

**Applies to:** `long`, `double` types

**Value:** Decimal between 0.0 and 1.0 (percentage)

```yaml
fields:
  # CPU usage changes by max 10% between events
  - name: cpu.percent
    fuzziness: 0.1
    range:
      min: 0
      max: 100

  # Network bytes with 5% variation
  - name: network.bytes
    fuzziness: 0.05
```

**How it works:**

1. First value: `10.0`
2. With `fuzziness: 0.1`, next value is between `9.0` and `11.0`
3. If next value is `10.5`, following value is between `9.45` and `11.55`

**Combined with range:** Values stay within range bounds even with fuzziness.

### `cardinality`

Limit the number of unique values generated for a field.

**Applies to:** All field types

```yaml
fields:
  # Only 100 unique host names
  - name: host.name
    cardinality: 100

  # Only 10 unique IP addresses
  - name: source.ip
    cardinality: 10

  # Combined with enum (cardinality limited by enum size)
  - name: region
    enum: ["us-east-1", "us-west-2", "eu-west-1"]
    cardinality: 3
```

**Important:** Cardinality may not be respected if:
- Not enough events are generated
- Other settings prevent it (e.g., small enum list)
- Range is too narrow

### `counter`

Generate ever-increasing values (monotonic counters).

**Applies to:** `long`, `double` types

```yaml
fields:
  # Unbounded counter
  - name: request.count
    counter: true

  # Counter with controlled growth (fuzziness)
  - name: bytes.total
    counter: true
    fuzziness: 0.1  # Increases by max 10% each time
```

**Note:** Cannot combine `counter: true` with `range.min` or `range.max`.

### `counter_reset`

Configure how and when counters reset to zero.

**Applies to:** Fields with `counter: true`

#### Random Reset

Reset at random intervals:

```yaml
fields:
  - name: bytes.total
    counter: true
    counter_reset:
      strategy: "random"
```

#### Probabilistic Reset

Reset based on probability for each event:

```yaml
fields:
  - name: bytes.total
    counter: true
    counter_reset:
      strategy: "probabilistic"
      probability: 5  # 5% chance to reset each event
```

#### Reset After N Events

Reset after a specific number of events:

```yaml
fields:
  - name: bytes.total
    counter: true
    counter_reset:
      strategy: "after_n"
      reset_after_n: 100  # Reset every 100 events
```

### `period`

Generate dates within a duration from the current time.

**Applies to:** `date` type

```yaml
fields:
  # Last hour
  - name: "@timestamp"
    period: "-1h"

  # Next 24 hours
  - name: scheduled_at
    period: "24h"

  # Last 7 days
  - name: event.created
    period: "-168h"  # 7 * 24 hours
```

**Format:** Go duration string (e.g., `"1h"`, `"30m"`, `"24h"`, `"-1h"`)

**Note:** Cannot combine `period` with `range.from` or `range.to`.

### `object_keys`

Specify keys to generate in object-type fields.

**Applies to:** `object` type

```yaml
fields:
  # Define object keys
  - name: aws.dimensions.*
    object_keys:
      - TableName
      - Operation
      - Region

  # Configure individual keys
  - name: aws.dimensions.TableName
    enum: ["users", "orders", "products"]

  - name: aws.dimensions.Operation
    enum: ["GetItem", "PutItem", "Query", "Scan"]

  - name: aws.dimensions.Region
    cardinality: 5
```

**Note:** If `cardinality` is set on the parent object, it overrides cardinality on child keys.

## Complete Example

```yaml
# configs.yml - AWS DynamoDB metrics configuration
fields:
  # Fixed values
  - name: data_stream.type
    value: metrics
  - name: data_stream.dataset
    value: aws.dynamodb
  - name: data_stream.namespace
    value: default

  # Date configuration
  - name: "@timestamp"
    period: "-1h"

  # Cardinality for realistic distribution
  - name: aws.cloudwatch.namespace
    cardinality: 1000
  - name: host.name
    cardinality: 50

  # Numeric ranges with fuzziness
  - name: aws.dynamodb.metrics.AccountMaxReads.max
    fuzziness: 0.1
    range:
      min: 0
      max: 100

  - name: aws.dynamodb.metrics.AccountMaxTableLevelReads.max
    fuzziness: 0.05
    range:
      min: 0
      max: 50
    cardinality: 20

  # Object with specific keys
  - name: aws.dimensions.*
    object_keys:
      - TableName
      - Operation

  - name: aws.dimensions.TableName
    enum: ["users", "orders", "products", "inventory"]

  - name: aws.dimensions.Operation
    cardinality: 5

  # Counter with reset
  - name: aws.dynamodb.metrics.ConsumedReadCapacityUnits.sum
    counter: true
    fuzziness: 0.1
    counter_reset:
      strategy: "after_n"
      reset_after_n: 1000
```

## Common Patterns

### Simulate Multiple Hosts

```yaml
fields:
  - name: host.name
    cardinality: 100  # 100 unique hosts
  - name: host.ip
    cardinality: 100  # Match host count
```

### Weighted Distribution

```yaml
fields:
  - name: log.level
    # 70% INFO, 20% WARN, 10% ERROR
    enum: ["INFO", "INFO", "INFO", "INFO", "INFO", "INFO", "INFO", "WARN", "WARN", "ERROR"]
```

### Time-Series Metrics

```yaml
fields:
  - name: "@timestamp"
    period: "-1h"
  - name: cpu.percent
    range:
      min: 0
      max: 100
    fuzziness: 0.1  # Smooth changes
```

### Cumulative Counters

```yaml
fields:
  - name: network.bytes.total
    counter: true
    fuzziness: 0.05
    counter_reset:
      strategy: "after_n"
      reset_after_n: 1440  # Reset daily (assuming 1 event/minute)
```

### Correlated Fields

Use array values with index for correlation:

```yaml
fields:
  - name: region_index
    range:
      min: 0
      max: 2
    cardinality: 3
  - name: regions
    value: ["us-east-1", "us-west-2", "eu-west-1"]
  - name: endpoints
    value: ["api.us-east.example.com", "api.us-west.example.com", "api.eu.example.com"]
```

## See Also

- [Field Types](./field-types.md) - Supported Elasticsearch field types
- [Writing Templates](./writing-templates.md) - Use configured fields in templates
- [Cardinality](./cardinality.md) - Deep dive into cardinality
- [Dimensionality](./dimensionality.md) - Configure object fields
