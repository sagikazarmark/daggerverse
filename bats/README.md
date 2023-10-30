# [Bats](https://github.com/bats-core/bats-core)

**Bash Automated Testing System**

## Examples

### Go

```go
dag.Bats().
    .WithSource(dag.Host().Directory("."))
    .Run([]string{"test.bats"})
```

## To Do

- [ ] Custom container with additional tools installed
- [ ] Better abstraction for running tests
