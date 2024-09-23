# Fields generation configuration

Fields generation configuration are applied to fields defined in Fields definition file to tweak how the data is generated.

They must be added to a file named `configs.yml` in the assets template folder of a data stream.

## Config entries definition

The config file is a yaml file consisting of root level `fields` object that's an array of config entry.

For each config entry the following fields are available:
- `name` *mandatory*: dotted path field, matching an entry in [Fields definition](./glossary.md#fields-definition)
- `fuzziness` *optional (`long` and `double` type only)*: when generating data you could want generated values to change in a known interval. Fuzziness allow to specify the maximum delta a generated value can have from the previous value (for the same field), as a delta percentage that will be applied below and above the previous value; value must be between 0.0 and 1.0, where 0 is 0% and 1 is 100%. When not specified there is no constraint on the generated values, boundaries will be defined by the underlying field type. For example, `fuzziness: 0.1`, assuming a `double` field type and with first value generated `10.`, will generate the second value in the range between `9.` and `11.`. Assuming the second value generated will be `10.5`, the third one will be generated in the range between `9.45` and `11.55`, and so on.
- `range` *optional (`long` and `double` type only)*: value will be generated between `min` and `max`. If `fuzziness` is defined, the value will be generated within a delta defined by `fuzziness` from the previous value. In any case (`fuzziness` or not) the value would not escape the `min`/`max` bounds.
- `range` *optional (`date` type only)*: value will be generated between `from` and `to`. Only one between `from` and `to` can be set, in this case the dates will be generated between `from`/`to` and `time.Now()`. Progressive order of the generated dates is always assured regardless the interval involving `from`, `to` and `time.Now()` is positive or negative. If both at least one of `from` or `to` and `period` settings are defined an error will be returned and the generator will stop. The format of the date must be parsable by the following golang date format: `2006-01-02T15:04:05.999999999-07:00`. 
- `cardinality` *optional*: exact number of different values to generate for the field; note that this setting may not be respected if not enough events are generated. For example, `cardinality: 1000` with `100` generated events would produce `100` different values, not `1000`. Similarly, the setting may not be respected if other settings prevents it. For example, `cardinality: 10` with an `enum` list of only 5 strings would produce `5` different values, not `10`. Or `cardinality: 10` for a `long` with `range.min: 1` and `range.max: 5` would produce `5` different values, not `10`. 
- `counter` *optional (`long` and  `double` type only)*: if set to `true` values will be generated only ever-increasing. If `fuzziness` is not defined, the positive delta from the previous value will be totally random and unbounded. For example, assuming `counter: true`, assuming a `int` field type and with first value generated `10.`, will generate the second value with any random value greater than `10`, like `11` or `987615243`. If `fuzziness` is defined, the value will be generated within a positive delta defined by `fuzziness` from the previous value. For example, `fuzziness: 0.1`, assuming `counter: true` , assuming a `double` field type and with first value generated `10.`, will generate the second value in the range between `10.` and `11.`. Assuming the second value generated will be `10.5`, the third one will be generated in the range between `10.5` and `11.55`, and so on. If both `counter: true` and at least one of `range.min` or `range.max` settings are defined an error will be returned and the generator will stop.
- `counter_reset` *optional (only applicable when `counter: true`)*: configures how and when the counter should reset. It has the following sub-fields:
  - `strategy` *mandatory*: defines the reset strategy. Possible values are:
      - `"random"`: resets the counter at random intervals.
      - `"probabilistic"`: resets the counter based on a probability.
      - `"after_n"`: resets the counter after a specific number of iterations.
  - `probability` *required when strategy is "probabilistic"*: an integer between 1 and 100 representing the percentage chance of reset for each generated value.
  - `reset_after_n` *required when strategy is "after_n"*: an integer specifying the number of values to generate before resetting the counter.
- `formatting_pattern` *optional (applicable to `string` type fields)*: a string that defines a pattern for generating formatted string values. The pattern can include static text and placeholders that will be replaced with random values. Multiple options can be provided, separated by `|`, from which one will be randomly selected for each generated value. Available placeholders are:
  - `{string}`: replaced with a random noun
  - `{ipv4}`: replaced with a random IPv4 address
  - `{ipv6}`: replaced with a random IPv6 address
  - `{port}`: replaced with a random port number between 1024 and 65535
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
