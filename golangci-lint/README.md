# GolangCI Lint

[Daggerverse](https://daggerverse.dev/mod/github.com/sagikazarmark/daggerverse/golangci-lint)
![Dagger Version](https://img.shields.io/badge/dagger%20version-%3E=0.9.8-0f0f19.svg?style=flat-square)

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
