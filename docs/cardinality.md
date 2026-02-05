# Cardinality

Cardinality is the number of distinct values a field can have. Understanding and controlling cardinality is essential for generating realistic test data.

## What is Cardinality?

Cardinality refers to the uniqueness of values in a field:

| Field Type | Typical Cardinality | Example Values |
|------------|---------------------|----------------|
| `boolean` | 2 | `true`, `false` |
| `log.level` | 4-5 | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `http.status_code` | ~60 | `200`, `404`, `500`, etc. |
| `host.name` | 10-10,000 | Server names in your fleet |
| `request.id` | Millions | Unique per request |

**Low cardinality:** Few unique values (boolean, status codes)

**High cardinality:** Many unique values (IDs, timestamps)

## Why Cardinality Matters

### Testing Perspective

High cardinality fields stress test your system:
- **Indexing performance** - More unique values = larger inverted index
- **Aggregation performance** - More buckets in terms aggregations
- **Memory usage** - Field data cache grows with cardinality
- **Storage** - Less compression for high cardinality fields

### Realism Perspective

Real observability data has specific cardinality patterns. For example, a Kubernetes cluster with 100 nodes, 1000 pods, and 50 namespaces would have:
- `kubernetes.node.name`: cardinality of 100
- `kubernetes.pod.name`: cardinality of 1000
- `kubernetes.namespace`: cardinality of 50

Generating data with wrong cardinality produces unrealistic test results.

## Configuring Cardinality

Set cardinality in your `configs.yml`:

```yaml
fields:
  - name: host.name
    cardinality: 100

  - name: source.ip
    cardinality: 50

  - name: user.id
    cardinality: 1000
```

### With Enums

Cardinality is limited by enum size:

```yaml
fields:
  - name: log.level
    enum: ["INFO", "WARN", "ERROR"]
    cardinality: 100  # Only 3 values possible
```

### With Ranges

Cardinality is limited by range:

```yaml
fields:
  - name: priority
    range:
      min: 1
      max: 5
    cardinality: 100  # Only 5 values possible
```

## Examples

### Simulating a Server Fleet

```yaml
fields:
  - name: host.name
    cardinality: 50
  - name: host.ip
    cardinality: 50
```

### Simulating Kubernetes

```yaml
fields:
  - name: kubernetes.node.name
    cardinality: 20
  - name: kubernetes.pod.name
    cardinality: 200
  - name: kubernetes.namespace
    cardinality: 10
```

## Best Practices

1. **Match real-world patterns** - Research actual cardinalities in production
2. **Consider event count** - Cardinality is limited by events generated
3. **Use correlated cardinalities** - Related fields should be consistent
4. **Test both extremes** - Generate low and high cardinality datasets
5. **Document assumptions** - Add comments explaining choices

## See Also

- [Fields Configuration](./fields-configuration.md) - All configuration options
- [Dimensionality](./dimensionality.md) - Field count in datasets
