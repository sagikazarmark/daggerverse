# Go

[Daggerverse](https://daggerverse.dev/mod/github.com/sagikazarmark/daggerverse/go)

**Yet another Dagger module for Go**

## Examples

### Go

```go
dag.Golang().
    .WithSource(dag.Host().Directory("."))
    .Exec([]string{"go", "build"})
```

### Shell

Run the following command to see the command line interface:

```shell
dagger call -m "github.com/sagikazarmark/daggerverse/go@main" --help
```

## To Do

- [ ] Add more tools
- [x] Add cache mounts
- [x] Add environment variables
- [ ] Add more examples
