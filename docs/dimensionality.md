# Dimensionality

The dimensionality of a dataset is the number of fields that it has. **High dimensionality** datasets have many attributes.

Dimensionality plays a role in `array` and `object` type fields.

Dimensionality is valuable to test storage consumption, in particular for metrics, where the data point itself is small compared to the metadata enriching it.

At the moment there is limited support for configuring dimensionality. When a field of type `object` is using `object_keys` is possible to configure the specific keys within the object itself.
Note that the `name: object_keys.*` configurations are not mandatory and if missing the field will be treated as having an empty config.

All unconfigured keys will be randomly generated.

For example:

```
- name: object_field
  object_keys:
    - a_key
    - another_key
- name: object_keys.a_key
  enum: ["a_value", "another_value"]
- name: object_keys.another_key
  cardinality: 2
```
