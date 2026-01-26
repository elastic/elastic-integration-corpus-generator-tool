# Dimensionality

Dimensionality refers to the number of fields in a dataset. High dimensionality datasets have many attributes, which is common in observability data.

## What is Dimensionality?

In the context of this tool, dimensionality primarily affects:

- **Object fields** - How many keys are generated within an object
- **Storage consumption** - More fields = more storage
- **Query complexity** - More fields to search and aggregate

## Why Dimensionality Matters

Observability data often has high dimensionality. A single metric data point might be small, but the metadata enriching it can be substantial:

```json
{
  "value": 42.5,
  "host": {"name": "server-01", "ip": "192.168.1.1"},
  "cloud": {"provider": "aws", "region": "us-east-1", "instance": {"id": "i-abc123"}},
  "kubernetes": {"pod": {"name": "app-xyz"}, "namespace": "production"},
  "labels": {"app": "myapp", "env": "prod", "team": "platform"}
}
```

Testing with realistic dimensionality helps validate storage estimates and query performance.

## Configuring Dimensionality

### Object Keys

Use `object_keys` to specify exact keys for object fields:

```yaml
fields:
  - name: labels.*
    object_keys:
      - app
      - env
      - team
      - version
```

### Configuring Individual Keys

Each object key can have its own configuration:

```yaml
fields:
  - name: aws.dimensions.*
    object_keys:
      - TableName
      - Operation
      - Region

  - name: aws.dimensions.TableName
    enum: ["users", "orders", "products"]

  - name: aws.dimensions.Operation
    enum: ["GetItem", "PutItem", "Query", "Scan"]

  - name: aws.dimensions.Region
    cardinality: 5
```

## Example

```yaml
fields:
  - name: metadata.*
    object_keys:
      - app
      - environment
      - version

  - name: metadata.app
    enum: ["frontend", "backend", "worker"]

  - name: metadata.environment
    enum: ["dev", "staging", "prod"]

  - name: metadata.version
    cardinality: 10
```

**Output:**
```json
{
  "metadata": {
    "app": "backend",
    "environment": "prod",
    "version": "v2.3.1"
  }
}
```

## Notes

- If `object_keys` is not specified, random keys are generated
- Child key configurations (`metadata.app`) are optional
- Parent object `cardinality` overrides child key cardinality

## See Also

- [Fields Configuration](./fields-configuration.md) - All configuration options
- [Field Types](./field-types.md) - Object type documentation
- [Cardinality](./cardinality.md) - Value uniqueness
