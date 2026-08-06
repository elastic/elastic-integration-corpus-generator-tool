# Glossary

A comprehensive reference of terms used in this tool and the Elastic ecosystem.

## Table of Contents

- [A](#a)
- [C](#c)
- [D](#d)
- [E](#e)
- [F](#f)
- [G](#g)
- [I](#i)
- [N](#n)
- [P](#p)
- [S](#s)
- [T](#t)

---

## A

### Agent
See [Elastic Agent](#elastic-agent).

---

## C

### Cardinality
The number of distinct values a field can have. Low cardinality fields (like `boolean`) have few unique values, while high cardinality fields (like `request_id`) can have millions. See [Cardinality Guide](./cardinality.md).

### Config File
A YAML file (`configs.yml`) that customizes how field values are generated. Controls ranges, cardinality, enums, and other generation parameters. See [Fields Configuration](./fields-configuration.md).

### Corpus
A collection of generated synthetic data events. The output of this tool is a corpus file in NDJSON format.

### Counter
A field configuration that generates ever-increasing values, simulating cumulative metrics like total bytes transferred.

---

## D

### Data Schema
One of four stages of data transformation in the Elastic pipeline (A, B, C, D). See [Data Schemas](./data-schemas.md).

### Data Stream
A logical grouping of time-series data in Elasticsearch, defined by type, dataset, and namespace. Part of the Elastic data model.

### Dataset
A component of a [Data Stream](#data-stream) that describes the ingested data structure. In this tool, `dataset` refers to the specific data you generate (e.g., `dynamodb`, `pod`).

### Dimensionality
The number of fields in a dataset. High dimensionality datasets have many attributes. See [Dimensionality Guide](./dimensionality.md).

---

## E

### Elastic Agent
The unified agent for collecting observability data (logs, metrics, traces) and shipping to Elasticsearch.

### Elastic Package Registry (EPR)
The central repository for Elastic integration packages. The `generate` command downloads field definitions from EPR.

### Enum
A configuration option that restricts field values to a predefined list of allowed values.

---

## F

### Field
A single data attribute with a name and type, defined in `fields.yml`.

### Field Type
The Elasticsearch data type for a field (e.g., `keyword`, `long`, `date`). See [Field Types](./field-types.md).

### Fields Definition
A YAML file (`fields.yml`) that defines the fields and their types for data generation.

### Fields Generation Configuration
See [Config File](#config-file).

### Fuzziness
A configuration option that controls how much consecutive values can change, creating smoother data progressions.

---

## G

### GoText
A template engine using Go's `text/template` package with Sprig helpers. More flexible but slower than [Placeholder](#placeholder).

---

## I

### Ingest Pipeline
An Elasticsearch feature that processes documents before indexing, transforming Schema B data into Schema C.

### Integration
An [Elastic integration package](https://www.elastic.co/integrations) that provides pre-built data collection for specific technologies (AWS, Kubernetes, etc.).

---

## N

### NDJSON
Newline-Delimited JSON. The output format of this tool, where each line is a valid JSON document. Compatible with Elasticsearch Bulk API.

---

## P

### Package
See [Integration](#integration).

### Period
A configuration option for date fields that specifies a duration from the current time for generating timestamps.

### Placeholder
A high-performance template engine that only supports simple variable substitution. Faster than [GoText](#gotext) but less flexible.

---

## S

### Schema A
Raw data format from the original source (log lines, API responses). See [Data Schemas](./data-schemas.md).

### Schema B
JSON data after Elastic Agent processing, before ingest pipeline. See [Data Schemas](./data-schemas.md).

### Schema C
Data after ingest pipeline processing, as stored in Elasticsearch. See [Data Schemas](./data-schemas.md).

### Schema D
Data as seen at query time, potentially including runtime fields. See [Data Schemas](./data-schemas.md).

### Seed
A random seed value for reproducible data generation. Using the same seed produces identical output.

### Sprig
A library of template functions available in GoText templates. See [Template Helpers](./go-text-template-helpers.md).

---

## T

### Template
A file (`.tpl`) that defines the output format for generated data. Uses either [GoText](#gotext) or [Placeholder](#placeholder) engine.

### Template Engine
The system that processes templates. This tool supports two engines: `gotext` (flexible) and `placeholder` (fast).

### Tot Events
The total number of events to generate. Set to `0` for infinite generation.

---

## Quick Reference

| Term | Definition |
|------|------------|
| Cardinality | Number of unique values for a field |
| Corpus | Generated data output file |
| Dataset | Specific data type within an integration |
| Dimensionality | Number of fields in a dataset |
| EPR | Elastic Package Registry |
| Fuzziness | Value change constraint between events |
| GoText | Flexible template engine |
| Integration | Elastic integration package |
| NDJSON | Output format (newline-delimited JSON) |
| Placeholder | Fast template engine |
| Schema A/B/C/D | Data transformation stages |
| Seed | Random seed for reproducibility |

---

## See Also

- [Data Schemas](./data-schemas.md) - Detailed schema explanation
- [Field Types](./field-types.md) - Supported field types
- [Fields Configuration](./fields-configuration.md) - Configuration options
- [Writing Templates](./writing-templates.md) - Template creation guide
