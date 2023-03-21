# Writing templates

Templates are, as implied from their name, files that define the format of the generated data by this tool.

The tool is designed to leverage any template in the compatible format. This file will guide you through writing template files for any occurrence.

There are 2 supported template types: `placeholder` (fast) and `gotext` (flexible). We suggest to use `gotext` unless performances are **critical** to your workflow.

## How does templating works?

At its core, this tool works by loading field definitions from a file and a template. The template is rendered using information from:
- the field definitions file
- the (optional) field generation configurations

This operation has some caveats:
- generating data is a computing and memory intensive task; you can expect to face computing pressure more easily while memory usage depends on the choosen template type and the template itself; performances of this tool is taken in great consideration, for more information see [performances.md](./performances.md);
- the use case is generating fake but plausible data, a challenge for a random data generator, as it implies:
  - in observability dataset usually exhibit high value cardinality; generating high cardinality dataset is a goal, for more information see [cardinality.md](./cardinality.md);
  - on the opposite side, a data generator may produce data with high similarity, which will not exhibit a behaviour similar to real data; while we aim to increase data dissimilarity, is not a goal to have complete real like data generation;
  - another characteristic of datasets in observability is high field dimensionality; generating high dimensionality dataset is a goal, for more information see [dimensionality.md](./dimensionality.md);
- data generation may not be a real time operation executed together/just before using the generated data; we aim to provide corpus of generated data that can be replayed as if they were being emitted in real time with the goal of allowing creation of big and complex datasets.

## Folder structure

All templates are contained in `/assets/templates` folder. Within it there should be a folder for each Integration data stream templates are available from.

The folder MUST be named as: `<package name>.<data stream>`. For example: `apache.access`, `apache.error`, `aws.billing`, `aws.vpcflow`. For example the final folder where to place the files listed below will be `/assets/templates/apache.error/schema-a`.

As this tool aims to support different [schemas](./data-schemas.md) for generated data, files related to a schema should be put within a folder named after the schema they relate to: `schema-a`, `schema-b`, `schema-c`, `schema-d`.

Within these folder, these files should be added:
- (_mandatory_) `fields.yml`: a field definition file
- (_optional_) `configs.yml`: a field generation configuration file
- (_mandatory_) `gotext.tpl`: a `gotext` template file
- (_optional_) `placeholder.tpl`: a `placeholder` template file

## `fields.yml` - Field definitions

A `YAML` file containing field mapping definitions. Ideally this file is extracted from Integration packages, but there is no automation for doing so at the moment.

## `configs.yml` - Field generation configurations

A `YAML` file containing configurations for field mappings defined in `fields.yml`. Details on configurations are in [field-configurations.md](./field-configurations.md).

This file allows to tweak the randomness of the generated data.

## Template types

This tool supports multiple templates types so to adapt to different scenarios. Rendering a template requires using a template engine and they greatly differ for performances and capabilities.

2 template engines are currently supported to optimise for 2 use cases:
- `placeholder` engine: uncompromised performances, is ok to lose features to gain performances;
- `gotext` engine: performant (but less) and feature rich, to aid development.

### placeholder

This template type is the most performant in terms of throughput: use this type **only** if data generation speed is relevant for you and you can trade off on the provided randomness and customisation given by the fields and config definitions.
The format of the template is similar to the one in Go `text/template` package, only supporting the feature subset of placeholder replacement.
Here's a template sample:
```text
{{ .Field1 }}-{{ .Field2 }} ({{ .Field3 }})
```

### gotext

This template type is less performant in terms of throughput than `placeholder` (our benchmarks shows from 3x to 9x slower according to the scenario), it uses the go text/template package with a few added functions: prefer this type as it supports data generation customisation that cannot be achieved only by the fields and config definitions.

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
