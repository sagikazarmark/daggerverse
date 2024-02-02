# Go

[Daggerverse](https://daggerverse.dev/mod/github.com/sagikazarmark/daggerverse/go)
![Dagger Version](https://img.shields.io/badge/dagger%20version-%3E=0.9.5-0f0f19.svg?style=flat-square)

**Yet another Dagger module for Go**

## Examples

### Go

```go
dag.Go().
    .WithSource(dag.CurrentModule().Source().Directory("."))
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
