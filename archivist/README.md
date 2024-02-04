# Archivist

[Daggerverse](https://daggerverse.dev/mod/github.com/sagikazarmark/daggerverse/archivist)
![Dagger Version](https://img.shields.io/badge/dagger%20version-%3E=0.9.8-0f0f19.svg?style=flat-square)

**Dagger module for creating and extracting archives.**

## Examples

### Go

This module implements the following interface:

```go
// Archiver archives a directory of files and returns a single archive.
type Archiver interface {
	DaggerObject

	Archive(ctx context.Context, name string, source *Directory) *File
}
```

Create the above interface in your code. Import this module in your main module.

Then you can use it like this:

```go
func (m *MyModule) archive(archiver Archiver, source *Directory) *File {
    return archiver.Archive(m, "archive", source)
}

func (m *MyModule) BuildArchive() *File {
    return m.archive(dag.Archivist().TarGz(), dag.Directory())
}
```

## To Do

- [ ] Add unarchive support
