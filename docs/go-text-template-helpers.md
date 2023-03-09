This file documents helper functions available when using the Go `text/template` template engine.

Within this template is possible to use **all** helpers from [`mastermings/sprig`](https://masterminds.github.io/sprig/).

<!-- helpers MUST be alphabetically sorted -->

# `awsAZFromRegion`

This helper accepts a string representing an AWS region (es. `us-east-1`) and returns a valid Availability Zone from that AWS region.

_NOTE_: Not all regions are supported at the moment (but can be added at need). For supported regions look [here](https://github.com/elastic/elastic-integration-corpus-generator-tool/blob/2c64e07461467aef4faacd5eb41efc3b0399c270/pkg/genlib/generator_with_text_template.go#L28-L30)

**Example**:

```text
{{ awsAZFromRegion "us-east-1" }}
```
```text
us-east-1a
```

# `timeDuration`

The helper accepts an `int64` and returns the equivalent `time.Duration`.

**Example**:

```text
{{$timeDuration := timeDuration 5000000000}}{{$timeDuration}}
```
```text
5s
```
