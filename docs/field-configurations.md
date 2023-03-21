# Field generation configurations

Field generation configurations are applied to fields defined in Fields definition file to tweak how the data is generated.

They must be added to a file named `configs.yml` in the assets template folder of a data stream.

## Config entries definition

The config file is a yaml file consisting of an array of config entry.

For each config entry the following fields are available:
- `name` *mandatory*: dotted path field, as in `fields.yml`
- `fuzziness` *optional (`long` and `double` type only)*: when generating data you could want generated values to change in a known interval. Fuzziness allow to specify the maximum delta a generated value can have from the previous value (for the same field), as a delta percentage; value must be between 0.0 and 1.0, where 0 is 0% and 1 is 100%. When not specified there is no constraint on the generated values, boundaries will be defined by the underlying field type
- `range` *optional (`long` and `double` type only)*: value will be generated between `min` and `max`
- `cardinality` *optional*: distribution of different values for the field, expressed as a ratio between a `numerator` and a `denominator`
- `object_keys` *optional (`object` type only)*: list of field names to generate in a object field type; if not specified a random number of field names will be generated in the object filed type
- `value` *optional*: hardcoded value to set for the field (any `cardinality` will be ignored)
- `enum` *optional (`keyword` type only)*: list of strings to randomly chose from a value to set for the field

If you have an `object` type field that you defined one or multiple `object_keys` for, you can reference them as a root level field with their own customisation. Beware that if a `cardinality` is set for the `object` type field, cardinality will be ignored for the children `object_keys` fields.

