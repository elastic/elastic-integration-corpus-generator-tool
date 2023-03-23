# elastic-integration-corpus-generator-tool
Command line tool used for generating events corpus dynamically given a specific integration

## Requirements
`make` CLI should be installed and available.

`git` CLI should be installed and available.

# Generate data from integration package fields
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


# Generate data from template
## Usage
```shell
$ ./elastic-integration-corpus-generator-tool generate-with-template -h
Generate a bulk request corpus given a template path and a fields definition path

Usage:
elastic-integration-corpus-generator-tool generate-with-template template-path fields-definition-path [flags]

Flags:
-c, --config-file string          path to config file for generator settings
-h, --help                        help for generate-with-template
-y, --template-type placeholder   either placeholder only or full `gotext` template (default "placeholder")
-t, --tot-size string             total size of the corpus to generate
```

#### Mandatory arguments
- template-path
- fields-definition-path

#### Mandatory flags
`--tot-size`

### Example
```shell
$ ./elastic-integration-corpus-generator-tool generate-with-template ./assets/templates/aws.vpcflow/vpcflow.gotext.log ./assets/templates/aws.vpcflow/vpcflow.fields.yml -t 20KB --config-file ./assets/templates/aws.vpcflow/vpcflow.conf.yml -y gotext -t 1000
File generated: /Users/andreaspacca/Library/Application Support/elastic-integration-corpus-generator-tool/corpora/1672731603-vpcflow.gotext.log
```

## Template types
### placeholder
This template type is the most performant in terms of throughput: use this type if data generation speed is relevant for you and you can trade off on the provided randomness and customisation given by the fields and config definitions.
The format of the template is similar to the one in go text/template package, only supporting the feature subset of placeholder replacement.
Here's a template sample:
```text
{{ .Field1 }}-{{ .Field2 }} ({{ .Field3 }})
```

Given the above template, in the fields definition file you'll have to define an entry for each field/placeholder, like in the following example (not that the `type` value is just for completeness, you can set the relevant value for your specific use case):
```yaml
- name: Field1
  type: long
- name: Field2
  type: ip
- name: Field3
  type: keyword
```

In the config file you can define all of only a subset of the fields used in the template, according to how you need to customise their behaviour, example:
```yaml
- name: Field1
  cardinality:
    numerator: 1
    denominator: 100
- name: Field3
  enum: ["value1", "value2"]
```

### gotext
This template type is less performant in terms of throughput from the above (our benchmarks shows from 3x to 9x slower according to the scenario), it uses the go text/template package with a few added functions: use this type if data generation customisation, that cannot be achieved only by the fields and config definitions, is relevant for you and you can trade off on speed.

#### "generate" function
The template provides a function named "generate" that accept the name of a field from the fields definition file as parameter: the output of this function is the random generated value for the field, respecting its definition and config:
```text
{{generate "Field1"}}
```

This is equivalent to the following when using the `placeholder` template type: 
```text
{{ .Field1 }}
```

#### sprig functions
The template loads the functions provided by sprig (https://masterminds.github.io/sprig/) with the exclusion of the functions are not guaranteed to evaluate to the same result for given input (https://github.com/Masterminds/sprig/blob/581758eb7d96ae4d113649668fa96acc74d46e7f/functions.go#L68-L95)

#### "timeDuration" function
The template provides a function named "timeDuration" that accept an int64 and return equivalent `time.Duration`, for example the following will render `5s`:
```text
{{$timeDuration := timeDuration 5000000000}}{{$timeDuration}} 
```

A sample template for AWS VPC Flow logs is the following:
```text
 {{generate "AccountID"}} {{generate "InterfaceID"}} {{generate "SrcAddr"}} {{generate "DstAddr"}} {{generate "SrcPort"}} {{generate "DstPort"}} {{generate "Protocol"}}{{ $packets := generate "Packets" }} {{ $packets }} {{mul $packets 15 }} {{$startOffset := generate "StartOffset" }}{{$startOffsetInSecond := mul -1 1000000000 $startOffset }}{{$startOffsetDuration := timeDuration $startOffsetInSecond}}{{$end := generate "End" }}{{$start := $end.Add $startOffsetDuration}}{{$start.Format "2006-01-02T15:04:05.999999Z07:00" }} {{$end.Format "2006-01-02T15:04:05.999999Z07:00"}} {{generate "Action"}}{{ if eq $packets 0 }} NODATA {{ else }} {{generate "LogStatus"}} {{ end }}
```

Alongside the following fields' definition:
```yaml
- name: Version
  type: long
- name: AccountID
  type: long
- name: InterfaceID
  type: keyword
  example: eni-1235b8ca123456789
- name: SrcAddr
  type: ip
- name: DstAddr
  type: ip
- name: SrcPort
  type: long
- name: DstPort
  type: long
- name: Protocol
  type: long
- name: Packets
  type: long
- name: End
  type: date
- name: StartOffset
  type: long
- name: Action
  type: keyword
- name: LogStatus
  type: keyword
```

And the following config file content:
```yaml
- name: Version
  value: 2
- name: AccountID
  value: 123456789012
- name: InterfaceID
  cardinality:
    numerator: 1
    denominator: 100
- name: SrcAddr
  cardinality:
    numerator: 1
    denominator: 1000
- name: DstAddr
  cardinality:
    numerator: 1
    denominator: 10
- name: SrcPort
  range:
    min: 0
    max: 65535
- name: DstPort
  range:
    min: 0
    max: 65535
  cardinality:
    numerator: 1
    denominator: 10
- name: Protocol
  range:
    min: 1
    max: 256
- name: Packets
  range:
    min: 1
    max: 1048576
- name: StartOffset
  range:
    min: 1
    max: 60
- name: Action
  enum: ["ACCEPT", "REJECT"]
- name: LogStatus
  enum: ["OK", "SKIPDATA"]
```


# Config file
It is possible to tweak the randomness of the generated data through a [config file](./docs/field-configurations.md) provided by the `--config-file` flag.

