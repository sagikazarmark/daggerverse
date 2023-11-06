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
        fromSource(source: ".") {
            lint(document: "openapi.yaml") {
                stdout
            }
        }
    }
}
```

### Shell

Run the following command to see the command line interface:

```shell
dagger call -m "github.com/sagikazarmark/daggerverse/spectral@main" --help
```

## To Do

- [ ] Custom ruleset parameters
- [ ] Custom arguments
- [ ] Lint multiple documents
