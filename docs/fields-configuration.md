# Fields generation configuration

Fields generation configuration are applied to fields defined in Fields definition file to tweak how the data is generated.

They must be added to a file named `configs.yml` in the assets template folder of a data stream.

## Config entries definition

The config file is a yaml file consisting of root level `fields` object that's an array of config entry.

For each config entry the following fields are available:
- `name` *mandatory*: dotted path field, matching an entry in [Fields definition](./glossary.md#fields-definition)
- `fuzziness` *optional (`long` and `double` type only)*: when generating data you could want generated values to change in a known interval. Fuzziness allow to specify the maximum delta a generated value can have from the previous value (for the same field), as a delta percentage; value must be between 0.0 and 1.0, where 0 is 0% and 1 is 100%. When not specified there is no constraint on the generated values, boundaries will be defined by the underlying field type
- `range` *optional (`long` and `double` type only)*: value will be generated between `min` and `max`, eventually according to the defined `fuzziness`.
- `range` *optional (`date` type only)*: value will be generated between `from` and `to`. Only one between `from` and `to` can be set, in this case the dates will be generated between `from`/`to` and `time.Now()`. Progressive order of the generated dates is always assured regardless the interval involving `from`, `to` and `time.Now()` is positive or negative. If both at least one of `from` or `to` and `period` settings are defined an error will be returned and the generator will stop. The format of the date must be parsable by the following golang date format: `2006-01-02T15:04:05.999999999-07:00`. 
- `cardinality` *optional*: exact number of different values to generate for the field; note that this setting may not be respected if not enough events are generated. Es `cardinality: 1000` with `100` generated events would produce `100` different values, not `1000`. Similarly, the setting may not be respected if other settings prevents it. Es `cardinality: 10` with an `enum` list of only 5 strings would produce `5` different values, not `10`. Or `cardinality: 10` for a `long` with `range.min: 1` and `range.max: 5` would produce `5` different values, not `10`. 
- `counter` *optional (`long` and  `double` type only)*: if set to `true` values will be generated only ever-increasing, eventually according to the defined `fuzziness`. If both `counter: true` and at least one of `range.min` or `range.max` settings are defined an error will be returned and the generator will stop.
- `period` *optional (`date` type only)*: values will be evenly generated between `time.Now()` and `time.Now().Add(period)`, where period is expressed as `time.Duration`. It accepts also a negative duration: in this case  values will be evenly generated between `time.Now().Add(period)` and `time.Now()`. If both `period` and at least one of `range.from` or `range.to` settings are defined an error will be returned and the generator will stop.
- `object_keys` *optional (`object` type only)*: list of field names to generate in a object field type; if not specified a random number of field names will be generated in the object filed type
- `value` *optional*: hardcoded value to set for the field (any `cardinality` will be ignored)
- `enum` *optional (`keyword` type only)*: list of strings to randomly chose from a value to set for the field (any `cardinality` will be applied limited to the size of the `enum` values)

If you have an `object` type field that you defined one or multiple `object_keys` for, you can reference them as a root level field with their own customisation. Beware that if a `cardinality` is set for the `object` type field, cardinality will be ignored for the children `object_keys` fields.

## Example configuration

```yaml
fields:
  - name: timestamp
    period: "1h"
  - name: lastSnapshot
    range:
      from: "2023-11-23T11:29:48-00:00"
      to: "2023-12-13T01:39:58-00:00"
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
  - name: aws.dynamodb.metrics.AccountProvisionedReadCapacityUtilization.avg
    fuzziness: 0.1
  - name: aws.cloudwatch.namespace
    cardinality: 1000
  - name: aws.dimensions.*
    object_keys:
      - TableName
      - Operation
  - name: data_stream.type
    value: metrics
  - name: data_stream.dataset
    value: aws.dynamodb
  - name: data_stream.namespace
    value: default
  - name: aws.dimensions.TableName
    enum: ["table1", "table2"]
  - name: aws.dimensions.Operation
    cardinality: 2
```

Related [fields definition](./writing-templates.md#fieldsyml---fields-definition)
```yaml
- name: timestamp
  type: date
- name: lastSnapshot
  type: date
- name: data_stream.type
  type: constant_keyword
- name: data_stream.dataset
  type: constant_keyword
- name: data_stream.namespace
  type: constant_keyword
- name: aws
  type: group
  fields:
    - name: dimensions
      type: group
      fields:
        - name: Operation
          type: keyword
        - name: TableName
          type: keyword
    - name: dynamodb
      type: group
      fields:
        - name: metrics
          type: group
          fields:
            - name: AccountProvisionedReadCapacityUtilization.avg
              type: double
            - name: AccountMaxReads.max
              type: long
            - name: AccountMaxTableLevelReads.max
              type: long
    - name: cloudwatch
      type: group
      fields:
        - name: namespace
          type: keyword
```
