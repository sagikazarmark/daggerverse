# [Spectral](https://stoplight.io/open-source/spectral)

**An open-source API style guide enforcer and linter.**

## Examples

### Go

```go
dag.Spectral().Lint(
    []*File{dag.CurrentModule().Source().File("openapi.json")},
    dag.CurrentModule().Source().File(".spectral.yaml"),
)
```

### GraphQL

```graphql
query test {
    spectral {
        lint(documents: ["openapi.yaml"], ruleset: ".spectral.yaml") {
            stdout
        }
    }
}
```

### Shell

Run the following command to see the command line interface:

```shell
dagger call -m "github.com/sagikazarmark/daggerverse/spectral@main" --help
```
