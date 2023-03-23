This file collects some common use cases for this tool.

# Generate schema-c data from integration package fields

To do this, use the `generate` command. This command targets a specific dataset within an integration package at a specific version.

You can pass a local Fields configuration file.

`go run main.go generate <package> <dataset> <version> --tot-size <quantity>`

`package`, `dataset` and `version` are mandatory. `--tot-size` is mandatory.

**Example**:

```shell
$ go run main.go generate aws dynamodb 1.14.0 -t 1000 --config-file config.yml
File generated: /path/to/corpora/1649330390-aws-dynamodb-1.14.0.ndjson
```

# Generate schema-b data from a template

To do this, use the `generate-with-template` command. This command targets a specific template within the `assets/templates` folder.

You can pass a local Fields configuration file.

`go run main.go generate-with-template <template-path> <fields-definition-path> --tot-size <quantity>`

`template-path` and `fields-definition-path` are mandatory. `--tot-size` is mandatory.

**Example**:

```shell
$ go run main.go generate-with-template ./assets/templates/aws.vpcflow/vpcflow.gotext.log ./assets/templates/aws.vpcflow/vpcflow.fields.yml -t 20KB --config-file ./assets/templates/aws.vpcflow/vpcflow.conf.yml -y gotext -t 1000
File generated: /Users/andreaspacca/Library/Application Support/elastic-integration-corpus-generator-tool/corpora/1672731603-vpcflow.gotext.log
```

