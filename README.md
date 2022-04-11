# elastic-integration-corpus-generator-tool
Command line tool used for generating events corpus dynamically given a specific integration

## Requirements
`make` CLI should be installed and available.

`git` CLI should be installed and available.

## Usage
```shell
$ ./elastic-integration-corpus-generator-tool generate -h
Generate a bulk request corpus for a given integration data stream downloaded from a package registry

Usage:
  elastic-integration-corpus-generator-tool generate integration data_stream version [flags]

Flags:
  -c, --config-file string                 path to config file for generator settings
  -h, --help                               help for generate
  -r, --package-registry-base-url string   base url of the package registry with schema (default "https://epr.elastic.co/")
  -t, --tot-size string                    total size of the corpus to generate
```

#### Mandatory arguments
- integration
- data_stream
- version

#### Mandatory flags
`--tot-size`

### Example
```shell
$ ./elastic-integration-corpus-generator-tool generate aws dynamodb 1.14.0 -t 1000 --config-file config.yml
File generated: /Users/andreaspacca/Library/Application Support/elastic-integration-corpus-generator-tool/corpora/1649330390-aws-dynamodb-1.14.0.ndjson
```


### Config file
It is possible to tweak the randomness of the generated data through a config file provivde by the `--config-file` flag

##### Sample config
```yaml
- name: aws.dynamodb.metrics.AccountMaxReads.max
  fuzziness: 10
  range: 100
- name: aws.dynamodb.metrics.AccountMaxTableLevelReads.max
  fuzziness: 5
  range: 50
  cardinality: 50
- name: aws.dynamodb.metrics.AccountProvisionedReadCapacityUtilization.avg
  fuzziness: 10
- name: aws.cloudwatch.namespace
  cardinality: 1
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
  cardinality: 1000
- name: aws.dimensions.Operation
  cardinality: 500
```

#### Config entries definition
The config file is a yaml file consisting of an array of config entry.
For each config entry the following fileds are available
- `name` *mandatory*: dotted path field
- `fuzziness` *optional (`long` and `double` type only)*: delta from the previous generated value for the same field
- `range` *optional (`long` and `double` type only)*: value will be generated between 0 and range
- `cardinality` *optional*: per-mille distribution of different values for the field
- `object_keys` *optional (`object` type only)*: list of field names to generate in a object field type. if not specified a random number of field names will be generated in the object filed type.
- `value` *optional*: hardcoded value to set for the field (any `cardinality` will be ignored)

If you have an `object` type field that you defined one or multiple `object_keys` for, you can reference them as a root level field with their own customisation. Beware that if a `cardinality` is set for the `object` type field, cardinality will be ignored for the children `object_keys` fields.