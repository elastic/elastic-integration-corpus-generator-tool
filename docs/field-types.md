# Field Types

This document describes all Elasticsearch field types supported by the corpus generator and how values are generated for each type.

## Table of Contents

- [Overview](#overview)
- [String Types](#string-types)
- [Numeric Types](#numeric-types)
- [Date Type](#date-type)
- [Boolean Type](#boolean-type)
- [IP Type](#ip-type)
- [Geo Point Type](#geo-point-type)
- [Complex Types](#complex-types)
- [Type Reference Table](#type-reference-table)

## Overview

Field types are specified in the `fields.yml` file:

```yaml
- name: my_field
  type: keyword
```

The generator creates appropriate random values based on the field type. You can further customize generation using [Fields Configuration](./fields-configuration.md).

## String Types

### `keyword`

Fixed-length string values, typically used for structured content like IDs, status codes, or tags.

**Default Behavior:** Generates random word combinations.

```yaml
- name: status
  type: keyword
```

**Configuration Options:**
- `enum`: List of allowed values
- `cardinality`: Number of unique values
- `value`: Fixed value

**Example Output:** `"eloquent-tiger"`, `"brave-mountain"`

### `constant_keyword`

A keyword that has the same value for all documents in an index.

**Default Behavior:** Uses the field's `example` value or generates a random word.

```yaml
- name: data_stream.type
  type: constant_keyword
```

**Configuration Options:**
- `value`: Fixed value (recommended)

**Example Output:** `"metrics"`, `"logs"`

## Numeric Types

### Integer Types

| Type | Range | Bytes |
|------|-------|-------|
| `byte` | -128 to 127 | 1 |
| `short` | -32,768 to 32,767 | 2 |
| `integer` | -2^31 to 2^31-1 | 4 |
| `long` | -2^63 to 2^63-1 | 8 |
| `unsigned_long` | 0 to 2^63-1* | 8 |

*Note: Currently generates values up to 2^63-1

**Default Behavior:** Generates random values within type bounds.

```yaml
- name: http.response.status_code
  type: integer
- name: bytes_transferred
  type: long
```

**Configuration Options:**
- `range.min`, `range.max`: Value bounds
- `fuzziness`: Maximum change between consecutive values (0.0-1.0)
- `cardinality`: Number of unique values
- `counter`: Generate ever-increasing values
- `counter_reset`: Reset strategy for counters

**Example Output:** `200`, `1048576`, `-42`

### Floating-Point Types

| Type | Precision | Description |
|------|-----------|-------------|
| `half_float` | 16-bit | Low precision, memory efficient |
| `float` | 32-bit | Single precision |
| `double` | 64-bit | Double precision |
| `scaled_float` | varies | Stored as long with scaling factor |

**Default Behavior:** Generates random decimal values.

```yaml
- name: cpu.percent
  type: double
- name: temperature
  type: float
```

**Configuration Options:**
- `range.min`, `range.max`: Value bounds
- `fuzziness`: Maximum change between consecutive values (0.0-1.0)
- `cardinality`: Number of unique values
- `counter`: Generate ever-increasing values

**Example Output:** `0.75`, `98.6`, `3.14159`

## Date Type

### `date`

Timestamp values in ISO 8601 format.

**Default Behavior:** Generates timestamps near the current time.

```yaml
- name: "@timestamp"
  type: date
- name: event.created
  type: date
```

**Configuration Options:**
- `period`: Duration from now (e.g., `"1h"`, `"-24h"`)
- `range.from`, `range.to`: Date range bounds

**Output Format:** `2024-01-15T10:30:00.000000Z`

**Example Configuration:**

```yaml
# Generate dates in the last hour
fields:
  - name: "@timestamp"
    period: "-1h"

# Generate dates in a specific range
fields:
  - name: event.created
    range:
      from: "2024-01-01T00:00:00-00:00"
      to: "2024-01-31T23:59:59-00:00"
```

## Boolean Type

### `boolean`

True or false values.

**Default Behavior:** Randomly generates `true` or `false`.

```yaml
- name: enabled
  type: boolean
- name: event.success
  type: boolean
```

**Configuration Options:**
- `value`: Fixed value (`true` or `false`)

**Example Output:** `true`, `false`

## IP Type

### `ip`

IPv4 or IPv6 addresses.

**Default Behavior:** Generates random IPv4 addresses.

```yaml
- name: source.ip
  type: ip
- name: destination.ip
  type: ip
```

**Configuration Options:**
- `cardinality`: Number of unique IP addresses

**Example Output:** `"192.168.1.100"`, `"10.0.0.50"`

## Geo Point Type

### `geo_point`

Geographic coordinates (latitude/longitude).

**Default Behavior:** Generates random valid coordinates.

```yaml
- name: location
  type: geo_point
- name: source.geo.location
  type: geo_point
```

**Output Format:** `{"lat": 40.7128, "lon": -74.0060}`

**Example Output:** `{"lat": 37.7749, "lon": -122.4194}`

## Complex Types

### `object`

Nested JSON objects with dynamic keys.

**Default Behavior:** Generates objects with random keys and values.

```yaml
- name: metadata
  type: object
- name: labels
  type: object
```

**Configuration Options:**
- `object_keys`: List of specific keys to generate

**Example Configuration:**

```yaml
fields:
  - name: aws.dimensions.*
    object_keys:
      - TableName
      - Operation
  - name: aws.dimensions.TableName
    enum: ["users", "orders", "products"]
  - name: aws.dimensions.Operation
    cardinality: 5
```

**Example Output:** `{"TableName": "users", "Operation": "GetItem"}`

### `nested`

Array of objects, each indexed as a separate document.

**Default Behavior:** Same as `object`.

```yaml
- name: items
  type: nested
```

### `flattened`

Entire object mapped as a single field.

**Default Behavior:** Same as `object`.

```yaml
- name: labels
  type: flattened
```

## Type Reference Table

| Type | Category | Default Generation | Configurable |
|------|----------|-------------------|--------------|
| `boolean` | Boolean | Random true/false | value |
| `keyword` | String | Random words | enum, cardinality, value |
| `constant_keyword` | String | From example or random | value |
| `date` | Date | Near current time | period, range |
| `ip` | Network | Random IPv4 | cardinality |
| `geo_point` | Geo | Random coordinates | - |
| `byte` | Integer | -128 to 127 | range, fuzziness, counter |
| `short` | Integer | -32K to 32K | range, fuzziness, counter |
| `integer` | Integer | Full 32-bit range | range, fuzziness, counter |
| `long` | Integer | Full 64-bit range | range, fuzziness, counter |
| `unsigned_long` | Integer | 0 to 2^63-1 | range, fuzziness, counter |
| `half_float` | Float | Random decimal | range, fuzziness, counter |
| `float` | Float | Random decimal | range, fuzziness, counter |
| `double` | Float | Random decimal | range, fuzziness, counter |
| `scaled_float` | Float | Random decimal | range, fuzziness, counter |
| `object` | Complex | Random keys/values | object_keys |
| `nested` | Complex | Random keys/values | object_keys |
| `flattened` | Complex | Random keys/values | object_keys |

## Unsupported Types

The following Elasticsearch types are not currently supported:
- `text` (use `keyword` instead)
- `binary`
- `range` types (`integer_range`, `date_range`, etc.)
- `completion`
- `search_as_you_type`
- `token_count`
- `dense_vector`
- `sparse_vector`
- `rank_feature`
- `rank_features`
- `shape`
- `histogram`
- `aggregate_metric_double`

For unsupported types, the generator falls back to generating random word strings.

## See Also

- [Fields Configuration](./fields-configuration.md) - Configure how values are generated
- [Writing Templates](./writing-templates.md) - Use fields in templates
- [Cardinality](./cardinality.md) - Control value diversity
- [Dimensionality](./dimensionality.md) - Configure object fields
