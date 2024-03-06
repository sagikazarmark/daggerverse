# GolangCI Lint

[Daggerverse](https://daggerverse.dev/mod/github.com/sagikazarmark/daggerverse/golangci-lint)

## Examples

### Go

```go
dag.GolangciLint().
    Run(dag.CurrentModule().Source().Directory("."))
```

### Shell

Run the following command to see the command line interface:

```shell
dagger call -m "github.com/sagikazarmark/daggerverse/golangci-lint@main" --help
```

## To Do
