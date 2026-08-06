# Data Schemas

Understanding data schemas is essential for generating the right type of test data. This guide explains the four data schemas in the Elastic data collection pipeline.

## Table of Contents

- [Overview](#overview)
- [Schema A: Raw Data](#schema-a-raw-data)
- [Schema B: Agent-Processed Data](#schema-b-agent-processed-data)
- [Schema C: Post-Ingest Data](#schema-c-post-ingest-data)
- [Schema D: Query-Time Data](#schema-d-query-time-data)
- [Choosing the Right Schema](#choosing-the-right-schema)
- [Schema Comparison](#schema-comparison)

## Overview

When collecting data with Elastic Agent and shipping to Elasticsearch, data passes through four transformation stages, each with its own schema:

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐     ┌───────────────┐     ┌─────────┐
│ Data Source │────▶│ Elastic Agent│────▶│ Ingest Pipeline │────▶│ Elasticsearch │────▶│  Query  │
│             │     │              │     │                 │     │               │     │         │
│  Schema A   │     │   Schema B   │     │    Schema C     │     │   Schema D    │     │         │
└─────────────┘     └──────────────┘     └─────────────────┘     └───────────────┘     └─────────┘
```

## Schema A: Raw Data

**What it is:** The original format from the data source before any processing.

**Examples:**
- Log file lines
- HTTP API responses
- Syslog messages
- Raw metrics output
- CSV data

**Characteristics:**
- Unstructured or semi-structured
- Source-specific format
- No Elastic metadata
- May require parsing

**Use Case:** Testing the full pipeline from data collection through ingest.

**Example - AWS VPC Flow Logs:**

```
2 627286350134 eni-abc123 192.168.1.100 10.0.0.50 54321 443 6 150 2250 2024-01-15T10:30:00Z 2024-01-15T10:30:05Z ACCEPT OK
```

**Generate Schema A:**

```bash
./elastic-integration-corpus-generator-tool local-template aws vpcflow -t 1000 --schema a
```

## Schema B: Agent-Processed Data

**What it is:** JSON data after Elastic Agent processing, before the ingest pipeline.

**Examples:**
- Filebeat output
- Metricbeat output
- Any Elastic Agent data stream output

**Characteristics:**
- JSON format
- Contains Elastic metadata (`agent`, `ecs`, `data_stream`)
- Original data in `message` field (for logs)
- Structured fields from agent processing
- Ready for ingest pipeline

**Use Case:** Testing ingest pipelines without running Elastic Agent.

**Example - Kubernetes Pod Metrics:**

```json
{
  "@timestamp": "2024-01-15T10:30:00.000000Z",
  "kubernetes": {
    "pod": {
      "name": "demo-pod-123",
      "uid": "abc-123-def",
      "cpu": {
        "usage": {
          "nanocores": 1500000000
        }
      },
      "memory": {
        "usage": {
          "bytes": 536870912
        }
      }
    },
    "namespace": "production"
  },
  "agent": {
    "id": "agent-001",
    "name": "node-01",
    "type": "metricbeat",
    "version": "8.12.0"
  },
  "data_stream": {
    "type": "metrics",
    "dataset": "kubernetes.pod",
    "namespace": "default"
  }
}
```

**Generate Schema B:**

```bash
./elastic-integration-corpus-generator-tool local-template kubernetes pod -t 1000 --schema b
```

## Schema C: Post-Ingest Data

**What it is:** Data after ingest pipeline processing, as stored in Elasticsearch.

**Examples:**
- Parsed log messages with extracted fields
- Enriched data (GeoIP, user agent parsing)
- Transformed metrics

**Characteristics:**
- JSON format
- Fully processed and enriched
- All fields populated
- Ready for indexing
- Matches `fields.yml` in integration packages

**Use Case:** Performance testing, Rally tracks, testing visualizations.

**Example - Parsed Log:**

```json
{
  "@timestamp": "2024-01-15T10:30:00.000000Z",
  "message": "User login successful",
  "event": {
    "action": "login",
    "outcome": "success",
    "category": ["authentication"]
  },
  "user": {
    "name": "john.doe",
    "id": "12345"
  },
  "source": {
    "ip": "192.168.1.100",
    "geo": {
      "city_name": "San Francisco",
      "country_name": "United States"
    }
  }
}
```

**Generate Schema C:**

```bash
./elastic-integration-corpus-generator-tool generate aws dynamodb 1.28.3 -t 1000
```

## Schema D: Query-Time Data

**What it is:** Data as seen when querying Elasticsearch, potentially with runtime fields.

**Characteristics:**
- Same as Schema C in most cases
- May include runtime field calculations
- Represents what users see in Kibana

**Relationship:** Schema D = Schema C + Runtime Fields

**Note:** This tool does not directly generate Schema D, as runtime fields are computed at query time.

## Choosing the Right Schema

| Goal | Schema | Command |
|------|--------|---------|
| Test full data collection pipeline | A | `local-template ... --schema a` |
| Test ingest pipelines | B | `local-template ... --schema b` |
| Performance testing / Rally tracks | C | `generate ...` |
| Test Kibana dashboards | C | `generate ...` |
| Integration development | B or C | Depends on what you're testing |

### Decision Tree

```
What are you testing?
│
├── Data collection (Agent inputs)?
│   └── Use Schema A
│
├── Ingest pipelines?
│   └── Use Schema B
│
├── Elasticsearch performance?
│   └── Use Schema C
│
├── Kibana visualizations?
│   └── Use Schema C
│
└── Full end-to-end?
    └── Use Schema A
```

## Schema Comparison

| Aspect | Schema A | Schema B | Schema C | Schema D |
|--------|----------|----------|----------|----------|
| Format | Various | JSON | JSON | JSON |
| Source | Raw data | Agent output | ES stored | ES query |
| Metadata | None | Partial | Full | Full |
| Parsing | Required | Partial | Complete | Complete |
| Use case | Collection | Ingest | Storage | Query |

### Field Progression Example

**Schema A (Raw):**
```
192.168.1.100 - john [15/Jan/2024:10:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234
```

**Schema B (Agent-Processed):**
```json
{
  "@timestamp": "2024-01-15T10:30:00Z",
  "message": "192.168.1.100 - john [15/Jan/2024:10:30:00 +0000] \"GET /api/users HTTP/1.1\" 200 1234",
  "log": {"file": {"path": "/var/log/nginx/access.log"}},
  "agent": {"name": "server-01", "type": "filebeat"}
}
```

**Schema C (Post-Ingest):**
```json
{
  "@timestamp": "2024-01-15T10:30:00Z",
  "message": "192.168.1.100 - john [15/Jan/2024:10:30:00 +0000] \"GET /api/users HTTP/1.1\" 200 1234",
  "source": {"ip": "192.168.1.100"},
  "user": {"name": "john"},
  "http": {
    "request": {"method": "GET"},
    "response": {"status_code": 200, "bytes": 1234}
  },
  "url": {"path": "/api/users"},
  "agent": {"name": "server-01", "type": "filebeat"}
}
```

## Available Templates by Schema

| Integration | Dataset | Schema A | Schema B |
|-------------|---------|----------|----------|
| AWS | billing | - | Yes |
| AWS | ec2_logs | - | Yes |
| AWS | ec2_metrics | - | Yes |
| AWS | sqs | - | Yes |
| AWS | vpcflow | Yes | - |
| Kubernetes | container | - | Yes |
| Kubernetes | pod | - | Yes |

## See Also

- [Writing Templates](./writing-templates.md) - Create templates for any schema
- [Usage Guide](./usage.md) - Common use cases
- [CLI Reference](./cli-help.md) - Command options
