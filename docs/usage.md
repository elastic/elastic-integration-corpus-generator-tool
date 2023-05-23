This file collects some common use cases for this tool.

# Generate schema-c data from integration package fields

To do this, use the `generate` command. This command targets a specific dataset within an integration package at a specific version.

You can pass a local Fields generation configuration file.

`go run main.go generate <package> <dataset> <version> --tot-events <quantity>`

`package`, `dataset` and `version` are mandatory. `--tot-events` is not mandatory and in case it is not provided a single event will be generated. You can generate an infinite number of events expressly passing to the flag the value of `0`. `--now` is not mandatory and in case it is provided must be a string parsable according the following `time.Parse()` layout: `2006-01-02T15:04:05.999999Z07:00`. The value provided will be used as base `time.Now()` for `date` type fields (see [Fields generation configuration](./fields-configuration.md#config-entries-definition))

**Example**:

```shell
$ go run main.go generate aws dynamodb 1.14.0 -t 1000 --config-file config.yml
File generated: /path/to/corpora/1649330390-aws-dynamodb-1.14.0.ndjson
```

# Generate schema-b data from a template

To do this, use the `generate-with-template` command. This command targets a specific template, fields definition and fields generation configuration.

A body of templates, fields definition and fields generation configuration are already available in the `assets/templates` folder. You can rely on them or write your own ones. Please consider opening a PR adding your custom templates, fields definition and fields configuration to the existing body, if they belong to an integration, so that we can enrich the catalogue for everyone.
 within the `assets/templates` folder.

You can pass a local Fields generation configuration file.

`go run main.go generate-with-template <template-path> <fields-definition-path> --tot-events <quantity>`

`template-path` and `fields-definition-path` are mandatory. `--tot-events` is not mandatory and in case it is not provided a single event will be generated. You can generate an infinite number of events expressly passing to the flag the value of `0`. `--now` is not mandatory and in case it is provided must be a string parsable according the following `time.Parse()` layout: `2006-01-02T15:04:05.999999Z07:00`. The value provided will be used as base `time.Now()` for `date` type fields (see [Fields generation configuration](./fields-configuration.md#config-entries-definition))

**Example**:

```shell
$ go run main.go generate-with-template ./assets/templates/aws.vpcflow/schema-a/gotext.tpl ./assets/templates/aws.vpcflow/schema-a/fields.yml -t 1000 --config-file ./assets/templates/aws.vpcflow/schema-a/configs.yml -y gotext
File generated: /path/to/corpora/1684304483-gotext.tpl
```

