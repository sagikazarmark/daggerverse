# [Spectral](https://stoplight.io/open-source/spectral)

**An open-source API style guide enforcer and linter.**

## Examples

### Go

```go
dag.Spectral().
    .WithSource(dag.Host().Directory("."))
    .Lint("openapi.yaml")
```

### GraphQL

```graphql
query test {
    spectral {
        withSource(source: {directory: "."}) {
            lint(document: "openapi.yaml") {
                stdout
            }
        }
    }
}
```
