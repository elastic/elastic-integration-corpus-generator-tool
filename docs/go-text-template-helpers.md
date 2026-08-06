# GoText Template Helpers

This document describes all helper functions available when using the `gotext` template engine.

## Table of Contents

- [Overview](#overview)
- [Core Functions](#core-functions)
- [Sprig Functions](#sprig-functions)
- [Custom Functions](#custom-functions)
- [Examples](#examples)

## Overview

GoText templates use Go's `text/template` package enhanced with:

1. **Core function**: `generate` - Generate field values
2. **Sprig library**: 70+ utility functions for strings, math, dates, etc.
3. **Custom functions**: Domain-specific helpers like `awsAZFromRegion`

## Core Functions

### `generate`

Generate a random value for a field based on its type and configuration.

**Syntax:**
```text
{{generate "field_name"}}
```

**Parameters:**
- `field_name`: The dotted path to a field defined in `fields.yml`

**Returns:** The generated value (type depends on field type)

**Examples:**

```text
{{/* String field */}}
{{generate "host.name"}}

{{/* Numeric field */}}
{{generate "http.response.status_code"}}

{{/* Date field */}}
{{generate "@timestamp"}}

{{/* Store in variable */}}
{{- $timestamp := generate "@timestamp" -}}
{{$timestamp.Format "2006-01-02T15:04:05Z07:00"}}
```

## Sprig Functions

All functions from the [Sprig library](https://masterminds.github.io/sprig/) are available. Here are the most commonly used categories:

### String Functions

| Function | Description | Example |
|----------|-------------|---------|
| `trim` | Remove whitespace | `{{trim " hello "}}` |
| `upper` | Uppercase | `{{upper "hello"}}` |
| `lower` | Lowercase | `{{lower "HELLO"}}` |
| `title` | Title case | `{{title "hello world"}}` |
| `replace` | Replace substring | `{{replace "hello" "l" "L" -1}}` |
| `split` | Split into list | `{{split "a,b,c" ","}}` |
| `join` | Join list | `{{join "-" (list "a" "b" "c")}}` |
| `substr` | Substring | `{{substr 0 5 "hello world"}}` |
| `contains` | Check contains | `{{contains "hello" "ell"}}` |
| `hasPrefix` | Check prefix | `{{hasPrefix "hello" "he"}}` |
| `hasSuffix` | Check suffix | `{{hasSuffix "hello" "lo"}}` |

### Math Functions

| Function | Description | Example |
|----------|-------------|---------|
| `add` | Addition | `{{add 1 2}}` |
| `sub` | Subtraction | `{{sub 5 2}}` |
| `mul` | Multiplication | `{{mul 3 4}}` |
| `div` | Division | `{{div 10 2}}` |
| `mod` | Modulo | `{{mod 10 3}}` |
| `max` | Maximum | `{{max 1 5 3}}` |
| `min` | Minimum | `{{min 1 5 3}}` |
| `floor` | Floor | `{{floor 3.7}}` |
| `ceil` | Ceiling | `{{ceil 3.2}}` |
| `round` | Round | `{{round 3.5 0}}` |

### Date Functions

| Function | Description | Example |
|----------|-------------|---------|
| `now` | Current time | `{{now}}` |
| `date` | Format date | `{{now \| date "2006-01-02"}}` |
| `dateModify` | Modify date | `{{now \| dateModify "-1h"}}` |
| `duration` | Parse duration | `{{duration "1h30m"}}` |
| `unixEpoch` | Unix timestamp | `{{now \| unixEpoch}}` |

### Type Conversion

| Function | Description | Example |
|----------|-------------|---------|
| `int` | To integer | `{{int "42"}}` |
| `int64` | To int64 | `{{int64 "42"}}` |
| `float64` | To float64 | `{{float64 "3.14"}}` |
| `toString` | To string | `{{toString 42}}` |
| `toJson` | To JSON | `{{toJson .data}}` |

### List Functions

| Function | Description | Example |
|----------|-------------|---------|
| `list` | Create list | `{{list "a" "b" "c"}}` |
| `first` | First element | `{{first (list 1 2 3)}}` |
| `last` | Last element | `{{last (list 1 2 3)}}` |
| `index` | Get by index | `{{index (list "a" "b") 1}}` |
| `append` | Append | `{{append (list 1 2) 3}}` |
| `concat` | Concatenate | `{{concat (list 1) (list 2)}}` |
| `len` | Length | `{{len (list 1 2 3)}}` |

### Dictionary Functions

| Function | Description | Example |
|----------|-------------|---------|
| `dict` | Create dict | `{{dict "key" "value"}}` |
| `get` | Get value | `{{get $dict "key"}}` |
| `set` | Set value | `{{set $dict "key" "value"}}` |
| `keys` | Get keys | `{{keys $dict}}` |
| `values` | Get values | `{{values $dict}}` |

### Conditional Functions

| Function | Description | Example |
|----------|-------------|---------|
| `default` | Default value | `{{default "N/A" .value}}` |
| `empty` | Check empty | `{{empty ""}}` |
| `coalesce` | First non-empty | `{{coalesce "" "default"}}` |
| `ternary` | Ternary operator | `{{ternary "yes" "no" true}}` |

For the complete list, see [Sprig Function Documentation](https://masterminds.github.io/sprig/).

## Custom Functions

### `awsAZFromRegion`

Get a random AWS Availability Zone for a given region.

**Syntax:**
```text
{{awsAZFromRegion "region"}}
```

**Parameters:**
- `region`: AWS region code (e.g., `us-east-1`)

**Returns:** A valid availability zone string (e.g., `us-east-1a`)

**Supported Regions:**

| Region | Availability Zones |
|--------|-------------------|
| `ap-east-1` | a, b, c |
| `ap-northeast-1` | a, c, d |
| `ap-northeast-2` | a, b, c, d |
| `ap-northeast-3` | a, b, c |
| `ap-south-1` | a, b, c |
| `ap-southeast-1` | a, b, c |
| `ap-southeast-2` | a, b, c |
| `ca-central-1` | a, b, d |
| `eu-central-1` | a, b, c |
| `eu-north-1` | a, b, c |
| `eu-west-1` | a, b, c |
| `eu-west-2` | a, b, c |
| `eu-west-3` | a, b, c |
| `me-south-1` | a, b, c |
| `sa-east-1` | a, b, c |
| `us-east-1` | a, b, c, d, e, f |
| `us-east-2` | a, b, c |
| `us-west-1` | a, b |
| `us-west-2` | a, b, c, d |

**Example:**
```text
{{- $region := generate "aws.region" -}}
{"region": "{{$region}}", "availability_zone": "{{awsAZFromRegion $region}}"}
```

**Output:**
```json
{"region": "us-east-1", "availability_zone": "us-east-1c"}
```

**Note:** Returns `"NoAZ"` for unsupported regions.

## Examples

### Date Manipulation

Calculate start time from end time and duration:

```text
{{- $end := generate "end_time" -}}
{{- $durationSec := generate "duration_seconds" -}}
{{- $start := $end | dateModify (mul -1 $durationSec | int64 | duration) -}}
{"start": "{{$start.Format "2006-01-02T15:04:05Z07:00"}}", "end": "{{$end.Format "2006-01-02T15:04:05Z07:00"}}"}
```

### Conditional Output

Different output based on field value:

```text
{{- $packets := generate "packets" -}}
{{- $status := "" -}}
{{- if eq $packets 0 -}}
{{- $status = "NODATA" -}}
{{- else -}}
{{- $status = generate "log_status" -}}
{{- end -}}
{"packets": {{$packets}}, "status": "{{$status}}"}
```

### Calculated Fields

Derive values from generated fields:

```text
{{- $packets := generate "packets" -}}
{{- $bytesPerPacket := 15 -}}
{"packets": {{$packets}}, "bytes": {{mul $packets $bytesPerPacket}}, "kilobytes": {{div (mul $packets $bytesPerPacket) 1024}}}
```

### Working with Arrays

Use index to access array elements:

```text
{{- $idx := generate "type_index" -}}
{{- $types := generate "instance_types" -}}
{"instance_type": "{{index $types $idx}}"}
```

### String Manipulation

Parse and reformat timestamps:

```text
{{- $timestamp := generate "timestamp" -}}
{{- $formatted := $timestamp.Format "2006-01-02T15:04:05.999999Z07:00" -}}
{{- $parts := split ":" $formatted -}}
{"date": "{{index $parts 0}}", "time": "{{index $parts 1}}:{{index $parts 2}}"}
```

### AWS-Specific Template

Generate AWS resource data:

```text
{{- $region := generate "aws.region" -}}
{{- $az := awsAZFromRegion $region -}}
{{- $instanceId := generate "aws.ec2.instance_id" -}}
{"cloud": {"provider": "aws", "region": "{{$region}}", "availability_zone": "{{$az}}"}, "instance": {"id": "{{$instanceId}}"}}
```

### Complex JSON Structure

Build nested JSON with multiple fields:

```text
{{- $timestamp := generate "@timestamp" -}}
{{- $host := generate "host.name" -}}
{{- $cpu := generate "system.cpu.percent" -}}
{{- $memory := generate "system.memory.used.bytes" -}}
{"@timestamp": "{{$timestamp.Format "2006-01-02T15:04:05.999999Z07:00"}}", "host": {"name": "{{$host}}"}, "system": {"cpu": {"percent": {{$cpu}}}, "memory": {"used": {"bytes": {{$memory}}}}}}
```

## Tips

### Whitespace Control

Use `-` to trim whitespace:

```text
{{- $var := generate "field" -}}
```

### Variable Scope

Variables defined with `:=` are scoped to the current block:

```text
{{- $global := "value" -}}
{{- if true -}}
  {{- $local := "inner" -}}
{{- end -}}
{{/* $local is not accessible here */}}
```

### Error Handling

Use `default` for fallback values:

```text
{{default "unknown" (generate "optional_field")}}
```

### Performance

For maximum performance, use the `placeholder` template engine instead. GoText is 3-9x slower but much more flexible.

## See Also

- [Writing Templates](./writing-templates.md) - Complete template guide
- [Sprig Documentation](https://masterminds.github.io/sprig/) - Full Sprig reference
- [Go text/template](https://pkg.go.dev/text/template) - Go template documentation
- [Performance](./performances.md) - Template engine benchmarks
