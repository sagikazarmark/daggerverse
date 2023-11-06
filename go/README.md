# Go

**Yet another Dagger module for Go**

## Examples

### Go

```go
dag.Golang().
    .WithSource(dag.Host().Directory("."))
    .Exec([]string{"go", "build"})
```

## To Do

- [ ] Add more tools
- [x] Add cache mounts
- [ ] Add environment variables
- [ ] Add more examples
