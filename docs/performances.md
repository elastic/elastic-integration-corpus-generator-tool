# Performances

Performances while generating data are a key element, as they:
- allow generating big quantities of data
- allow generating data on demand
- allow re-generating data when needed

Achieving maximum performances has also some disadvantages, mainly around feature-completeness.

This tool aims to support uses cases where is needed to trade features for performances and use cases where this isn't needed, but without compromising performances too much (max 10x less performant).

This tool comes with a benchmark testsuite that can be run to evaluate new features performance impact.

## Benchmarks

In PR #40, where multiple templates support has been added, we run two different benchmark for each template engine:

    JSONContent: producing Schema C data for "endpoint process 8.2.0" integration
    VPCFlowLogs: producing Schema A data for aws vpc flow logs
    (beware the memory benchmark for Hero are misleading since they "happens" in a forked process)

We tested 3 different engines:
- `legacy`: the only generator available before #40
- `CustomTemplate`: what we refer now as `placeholder` template engine
- `TextTemplate`: what we refer now as `gotext` template engine

```
_GeneratorLegacyJSONContent: the original generator, generating from fields definitions for endpoint package v8.2.0 data stream "process"
_GeneratorCustomTemplateJSONContent-16: placeholder template with arbitrary JSON content
_GeneratorTextTemplateJSONContent-16: Go text/template with arbitraty JSON content
_GeneratorCustomTemplateVPCFlowLogs-16: placeholder template generating Schema A data for AWS VPCFlowLogs
_GeneratorTextTemplateVPCFlowLogs-16: Go text/template generating Schema A data for AWS VPCFlowLogs

Tests have been executed on a 16 Cores machine.

name                                    time/op
_GeneratorLegacyJSONContent-16          47.7µs ± 0%
_GeneratorCustomTemplateJSONContent-16  30.0µs ± 0%
_GeneratorTextTemplateJSONContent-16     281µs ± 0%
_GeneratorCustomTemplateVPCFlowLogs-16  1.09µs ± 0%
_GeneratorTextTemplateVPCFlowLogs-16    12.8µs ± 0%

name                                    alloc/op
_GeneratorLegacyJSONContent-16          3.82kB ± 0%
_GeneratorCustomTemplateJSONContent-16    432B ± 0%
_GeneratorTextTemplateJSONContent-16    48.3kB ± 0%
_GeneratorCustomTemplateVPCFlowLogs-16   64.0B ± 0%
_GeneratorTextTemplateVPCFlowLogs-16    2.32kB ± 0%

name                                    allocs/op
_GeneratorLegacyJSONContent-16            22.0 ± 0%
_GeneratorCustomTemplateJSONContent-16    14.0 ± 0%
_GeneratorTextTemplateJSONContent-16     2.23k ± 0%
_GeneratorCustomTemplateVPCFlowLogs-16    2.00 ± 0%
_GeneratorTextTemplateVPCFlowLogs-16      95.0 ± 0%

```

If you are curious how those benchmarks translate to time needed for generating dataset, we ran some test runs monitoring the execution times.
We generated directly from the built binaries 20GB of "aws dynamodb 1.28.3" Schema C data.

```
$ time ./gen-legacy generate aws dynamodb 1.28.3 -t 20GB
File generated: [...]/elastic-integration-corpus-generator-tool/corpora/1671594228-aws-dynamodb-1.28.3.ndjson

real	1m44.869s
user	1m6.599s
sys	0m37.354s


$ time ./gen-with-custom_template generate aws dynamodb 1.28.3 -t 20GB
File generated: [...]/elastic-integration-corpus-generator-tool/corpora/1671611719-aws-dynamodb-1.28.3.ndjson

real	1m34.968s
user	0m55.029s
sys	0m37.175s


$ time ./gen-with-text_template generate aws dynamodb 1.28.3 -t 20GB
File generated: [...]/elastic-integration-corpus-generator-tool/corpora/1671612518-aws-dynamodb-1.28.3.ndjson

real	6m50.022s
user	6m10.642s
sys	0m50.909s
```
