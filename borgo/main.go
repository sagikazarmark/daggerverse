// Borgo is a statically typed language that compiles to Go.
package main

import "dagger/borgo/internal/dagger"

type Borgo struct {
	// +private
	Container *dagger.Container

	// +private
	Std *dagger.Directory
}

func New(
	// +default="dc4f72d04e7a4e0498b25d1ddea411d83509be7f"
	// +optional
	commit string,
) *Borgo {
	source := dag.Git("https://github.com/borgo-lang/borgo").Commit(commit).Tree()

	binary := dag.Container().
		From("rust:latest").
		WithDirectory("/work", source).
		WithWorkdir("/work").
		WithExec([]string{"cargo", "build", "--release"}).
		File("/work/target/release/compiler")

	container := dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		WithExec([]string{"apk", "add", "gcc"}).
		WithFile("/usr/bin/borgo", binary)

	return &Borgo{
		Container: container,
		Std:       source.Directory("std"),
	}
}

func (m *Borgo) Compile(source *dagger.Directory) *dagger.Directory {
	return m.Container.
		WithWorkdir("/work").
		WithDirectory("/work", source).
		WithDirectory("/work/std", m.Std).
		WithExec([]string{"borgo", "build"}).
		Directory("/work")
}

func (m *Borgo) Terminal(source *dagger.Directory) *dagger.Container {
	return m.Container.
		WithWorkdir("/work").
		WithDirectory("/work", source).
		WithDirectory("/work/std", m.Std).
		Terminal()
}
