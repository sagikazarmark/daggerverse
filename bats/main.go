// Bash Automated Testing System
package main

import (
	"dagger/bats/internal/dagger"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "bats/bats"

type Bats struct {
	// +private
	Container *dagger.Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	version string,

	// Custom container to use as a base container.
	//
	// +optional
	container *dagger.Container,
) *Bats {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	return &Bats{
		Container: container,
	}
}

// Mount a source directory.
func (m *Bats) WithSource(
	// Source directory.
	source *dagger.Directory,
) *WithSource {
	const workdir = "/work"

	return &WithSource{
		Source: source,
		Bats:   m,
	}
}

// Run bats tests.
func (m *Bats) Run(
	// Arguments to pass to bats.
	args []string,

	// Source directory to mount.
	//
	// +optional
	source *dagger.Directory,
) *dagger.Container {
	if source != nil {
		return m.WithSource(source).Run(args)
	}

	return m.Container.WithExec(args)
}

type WithSource struct {
	// +private
	Source *dagger.Directory

	// +private
	Bats *Bats
}

// Run bats tests.
func (m *WithSource) Run(
	// Arguments to pass to bats.
	args []string,
) *dagger.Container {
	const workdir = "/work"

	return m.Bats.Container.
		WithWorkdir(workdir).
		WithMountedDirectory(workdir, m.Source).
		WithExec(append([]string{"bats"}, args...))
}
