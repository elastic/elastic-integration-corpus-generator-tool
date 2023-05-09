# Corpus

# Data schemas

Different data structures we observe in the end to end data collection flow for Elastic (Beats & Agent). See [Data Schemas](./data-schemas.md).

In the context of this tool it also refers to folder names containing template information for a specific data schema.

# Dataset

A dataset is a component of a [Data Stream](https://www.elastic.co/guide/en/fleet/master/data-streams.html). Is defined by an _integration_ and describes the ingested data and its structure.

In the context of this tool `dataset` refers to the name of the dataset you can generate data for. Data structure and definitions are part of the integration package the dataset belongs to.

# Fields definition

Datasets define data structure. Within them there are fields definition that describe the fields a dataset provides. 

In the context of this tool we can refer to:
- the fields definition within a dataset in an integration package
- a file named `fields.yml` that contains fields definition that is not (yet) part of a package

# Fields generation configuration

Complete randomness in generated data may not always be advisable. There may be relationships or constraints that must be expressed to create corpus that have life-like characteristics. Through the Fields generation configuration file, named `configs.yml`, is possible to specify these constraints.

See [Fields generation configuration](./fields-configuration.md).

# Integration

An [Elastic integration package](https://www.elastic.co/guide/en/integrations-developer/current/what-is-an-integration.html).

# Template

A file containing a template for a specific template engine.
