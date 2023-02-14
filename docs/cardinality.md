# Cardinality

The cardinality of a field refers to the number of distinct values that it can have.

For example: a `boolean` will have cardinality of 2, an `integer` 4294967295, a `version` field may have a cardinality of some dozens.

Low cardinality fields, like `boolean` or `version` in the example above do not pose a particular issue when observing your system.
This is the opposite for high cardinality fields, which due to their size create challenges in indexing, searching and visualising them.

We refer to **high cardinality** when fields cardinality is in the order of hundreds of thousands or millions.

Example of these values may be: request IDs, trace IDs, value of tags attached to compute instances (es a tag with 20 distinct values in a 5000 instances fleet).

These fields are extremely valuable as they allow to fine grain your search. From a testing point of view they allow to stress test a system.

## Field generation configuration

For these reasons one of the goals for this tool is to be able to generate high cardinality fields. An additional complexity we face is to generate plausible cardinality. 

Let's make an exmaple: we manage a fleet of 1000 Kubernetes nodes. Each node hosts 100 pods. Pods in Kubernetes are within a namespace. Let's say we want to test the use case of few namespaces with thousands of pods (i.e. 1:1000 ration). This is a valid scenario, but we may be interested in another use case: namespaces containing very few pods (i.e. 1:10 ratio).

To support generating dataset for both uses cases, is possible to specify a `cardinality` paramenter in the field generation configuration file to tweak generated data.
See [field-configurations.md](./field-configurations.md).

