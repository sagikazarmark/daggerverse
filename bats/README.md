# [Bats](https://github.com/bats-core/bats-core)

**Bash Automated Testing System**

## Examples

### Go

```go
dag.Bats().
    .WithSource(dag.Host().Directory("."))
    .Run([]string{"test.bats"})
```

### Shell

Run the following command to see the command line interface:

```shell
dagger call -m "github.com/sagikazarmark/daggerverse/spectral@main" --help
```

## To Do

- [ ] Custom container with additional tools installed
- [ ] Better abstraction for running tests
