# Writing Templates

Templates define the format of generated data. This guide covers everything you need to know to create effective templates.

## Table of Contents

- [Overview](#overview)
- [Template Engines](#template-engines)
- [Folder Structure](#folder-structure)
- [Required Files](#required-files)
- [GoText Templates](#gotext-templates)
- [Placeholder Templates](#placeholder-templates)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

At its core, this tool works by:

1. Loading field definitions from a `fields.yml` file
2. Loading optional configuration from a `configs.yml` file
3. Rendering a template using the field definitions and configuration
4. Outputting generated events

Templates control the **structure** of the output, while field definitions and configurations control the **values**.

## Template Engines

Two template engines are supported, optimized for different use cases:

| Engine | Performance | Features | Best For |
|--------|-------------|----------|----------|
| `gotext` | Good | Full Go text/template + Sprig helpers | Development, complex templates |
| `placeholder` | Excellent (3-9x faster) | Simple variable substitution only | Production, large datasets |

**Recommendation:** Use `gotext` unless performance is critical.

## Folder Structure

Templates are organized in `assets/templates/` by integration and schema:

```
assets/templates/
  <package>.<dataset>/
    schema-a/
      fields.yml        # Field definitions (required)
      configs.yml       # Field configuration (optional)
      gotext.tpl        # GoText template (required)
      placeholder.tpl   # Placeholder template (optional)
    schema-b/
      fields.yml
      configs.yml
      gotext.tpl
      placeholder.tpl
```

### Naming Convention

- Folder: `<package>.<dataset>` (e.g., `aws.vpcflow`, `kubernetes.pod`)
- Schema folders: `schema-a`, `schema-b`, `schema-c`, `schema-d`

## Required Files

### `fields.yml` - Field Definitions

Defines the fields and their Elasticsearch types:

```yaml
- name: timestamp
  type: date
- name: message
  type: keyword
- name: host.name
  type: keyword
  example: server-01
- name: http.status_code
  type: integer
- name: response.time
  type: double
- name: enabled
  type: boolean
- name: source.ip
  type: ip
- name: location
  type: geo_point
```

**Supported Types:**
- `boolean`
- `keyword`, `constant_keyword`
- `date`
- `ip`
- `double`, `float`, `half_float`, `scaled_float`
- `byte`, `short`, `integer`, `long`, `unsigned_long`
- `object`, `nested`, `flattened`
- `geo_point`

See [Field Types](./field-types.md) for complete documentation.

### `configs.yml` - Field Configuration (Optional)

Fine-tunes how values are generated:

```yaml
fields:
  - name: timestamp
    period: "-1h"
  - name: http.status_code
    enum: ["200", "201", "400", "404", "500"]
  - name: response.time
    range:
      min: 0.1
      max: 5.0
    fuzziness: 0.1
  - name: host.name
    cardinality: 100
```

See [Fields Configuration](./fields-configuration.md) for complete documentation.

## GoText Templates

GoText templates use Go's `text/template` package with additional helpers.

### Basic Syntax

```text
{{generate "field_name"}}
```

The `generate` function outputs a random value for the specified field based on its type and configuration.

### Example: JSON Output

```text
{"timestamp":"{{generate "timestamp"}}","host":{"name":"{{generate "host.name"}}"},"message":"{{generate "message"}}"}
```

### Example: Log Line Output

```text
[{{generate "timestamp"}}] {{generate "log.level"}} {{generate "host.name"}} - {{generate "message"}}
```

### Variables and Calculations

Store generated values in variables for reuse:

```text
{{- $timestamp := generate "timestamp" -}}
{{- $bytes := generate "bytes" -}}
{"timestamp":"{{$timestamp.Format "2006-01-02T15:04:05.999999Z07:00"}}","bytes":{{$bytes}},"kilobytes":{{div $bytes 1024}}}
```

### Conditional Logic

```text
{{- $status := generate "status" -}}
{{- if eq $status "error" -}}
{"level":"ERROR","status":"{{$status}}","alert":true}
{{- else -}}
{"level":"INFO","status":"{{$status}}","alert":false}
{{- end -}}
```

### Date Manipulation

```text
{{- $end := generate "end_time" -}}
{{- $duration := generate "duration_seconds" -}}
{{- $start := $end | dateModify (mul -1 $duration | int64 | duration) -}}
{"start":"{{$start.Format "2006-01-02T15:04:05Z07:00"}}","end":"{{$end.Format "2006-01-02T15:04:05Z07:00"}}"}
```

### Available Helpers

GoText templates include all [Sprig functions](https://masterminds.github.io/sprig/) plus custom helpers:

| Helper | Description | Example |
|--------|-------------|---------|
| `generate` | Generate field value | `{{generate "field_name"}}` |
| `awsAZFromRegion` | Get AWS AZ from region | `{{awsAZFromRegion "us-east-1"}}` |

See [Template Helpers](./go-text-template-helpers.md) for complete documentation.

## Placeholder Templates

Placeholder templates offer maximum performance with simple variable substitution.

### Syntax

```text
{{ .FieldName }}
```

**Note:** Field names use dots replaced with nothing and PascalCase.

### Example

For fields:
```yaml
- name: host.name
  type: keyword
- name: message
  type: keyword
```

Template:
```text
{{ .HostName }} - {{ .Message }}
```

### Limitations

- No conditional logic
- No variable storage
- No calculations
- No helper functions
- Field names must be valid Go identifiers

## Best Practices

### 1. Use Whitespace Control

Remove unwanted whitespace with `-`:

```text
{{- $var := generate "field" -}}
```

### 2. Match Output Format to Schema

- **Schema A**: Raw format (log lines, CSV, etc.)
- **Schema B**: JSON format (Elastic Agent output)
- **Schema C**: JSON format (post-ingest pipeline)

### 3. Validate JSON Output

For JSON templates, ensure valid JSON:

```text
{"field":"{{generate "field"}}","number":{{generate "number"}}}
```

Note: Strings need quotes, numbers don't.

### 4. Use Meaningful Field Names

```yaml
# Good
- name: kubernetes.pod.memory.usage.bytes
  type: long

# Avoid
- name: mem
  type: long
```

### 5. Document Complex Templates

Add comments explaining logic:

```text
{{- /* Calculate start time from end time and duration */ -}}
{{- $end := generate "end_time" -}}
{{- $duration := generate "duration" -}}
```

### 6. Test with Small Datasets First

```bash
./elastic-integration-corpus-generator-tool generate-with-template ./template.tpl ./fields.yml -t 10 -y gotext
```

## Examples

### AWS VPC Flow Logs (Schema A)

**fields.yml:**
```yaml
- name: Version
  type: long
- name: AccountID
  type: long
- name: InterfaceID
  type: keyword
- name: SrcAddr
  type: ip
- name: DstAddr
  type: ip
- name: SrcPort
  type: long
- name: DstPort
  type: long
- name: Protocol
  type: long
- name: Packets
  type: long
- name: Action
  type: keyword
- name: LogStatus
  type: keyword
- name: End
  type: date
- name: StartOffset
  type: long
```

**configs.yml:**
```yaml
fields:
  - name: Version
    value: 2
  - name: AccountID
    value: 627286350134
  - name: InterfaceID
    cardinality: 100
  - name: SrcAddr
    cardinality: 1000
  - name: DstAddr
    cardinality: 10
  - name: SrcPort
    range:
      min: 0
      max: 65535
  - name: DstPort
    range:
      min: 0
      max: 65535
    cardinality: 10
  - name: Action
    enum: ["ACCEPT", "REJECT"]
  - name: LogStatus
    enum: ["OK", "SKIPDATA"]
```

**gotext.tpl:**
```text
{{- $startOffset := generate "StartOffset" }}
{{- $end := generate "End" }}
{{- $start := $end | dateModify (mul -1 $startOffset | int64 | duration) }}
{{generate "Version"}} {{generate "AccountID"}} {{generate "InterfaceID"}} {{generate "SrcAddr"}} {{generate "DstAddr"}} {{generate "SrcPort"}} {{generate "DstPort"}} {{generate "Protocol"}}{{ $packets := generate "Packets" }} {{ $packets }} {{mul $packets 15 }} {{$start.Format "2006-01-02T15:04:05.999999Z07:00" }} {{$end.Format "2006-01-02T15:04:05.999999Z07:00"}} {{generate "Action"}} {{generate "LogStatus"}}
```

### Simple JSON Event (Schema B)

**fields.yml:**
```yaml
- name: timestamp
  type: date
- name: event.action
  type: keyword
- name: user.name
  type: keyword
- name: source.ip
  type: ip
```

**configs.yml:**
```yaml
fields:
  - name: timestamp
    period: "-24h"
  - name: event.action
    enum: ["login", "logout", "access", "modify", "delete"]
  - name: user.name
    cardinality: 50
```

**gotext.tpl:**
```text
{"@timestamp":"{{(generate "timestamp").Format "2006-01-02T15:04:05.999999Z07:00"}}","event":{"action":"{{generate "event.action"}}"},"user":{"name":"{{generate "user.name"}}"},"source":{"ip":"{{generate "source.ip"}}"}}
```

## Contributing Templates

When adding templates to the repository:

1. Create folder: `assets/templates/<package>.<dataset>/schema-<x>/`
2. Add required files: `fields.yml`, `gotext.tpl`
3. Add optional files: `configs.yml`, `placeholder.tpl`
4. Test thoroughly with various event counts
5. Open a PR with your contribution

## See Also

- [Field Types](./field-types.md) - Supported Elasticsearch field types
- [Fields Configuration](./fields-configuration.md) - Configure data generation
- [Template Helpers](./go-text-template-helpers.md) - Available helper functions
- [Data Schemas](./data-schemas.md) - Understanding A/B/C/D schemas
- [Performance](./performances.md) - Template engine benchmarks
