This file documents helper functions available when using the Go `text/template` template engine.

Within this template is possible to use **all** helpers from [`mastermings/sprig`](https://masterminds.github.io/sprig/).

<!-- helpers MUST be alphabetically sorted -->

# `timeDuration`

The helper accepts an `int64` and returns the equivalent `time.Duration`.

**Example**:

```text
{{$timeDuration := timeDuration 5000000000}}{{$timeDuration}}
```
```text
5s
```
