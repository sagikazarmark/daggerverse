package main

import (
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "bats/bats"

type Bats struct {
	// +private
	Ctr *Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	// +optional
	version string,

	// Custom image reference in "repository:tag" format to use as a base container.
	// +optional
	image string,

	// Custom container to use as a base container.
	// +optional
	container *Container,
) *Bats {
	var ctr *Container

	if version != "" {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	} else if image != "" {
		ctr = dag.Container().From(image)
	} else if container != nil {
		ctr = container
	} else {
		ctr = dag.Container().From(defaultImageRepository)
	}

	return &Bats{ctr}
}

func (m *Bats) Container() *Container {
	return m.Ctr
}

// Mount a source directory.
func (m *Bats) WithSource(
	// Source directory.
	src *Directory,
) *WithSource {
	const workdir = "/work"

	return &WithSource{
		&Bats{
			m.Ctr.
				WithWorkdir(workdir).
				WithMountedDirectory(workdir, src),
		},
	}
}

// Run bats tests.
func (m *Bats) Run(
	// Arguments to pass to bats.
	args []string,

	// Source directory to mount.
	// +optional
	src *Directory,
) *Container {
	if src != nil {
		return m.WithSource(src).Run(args)
	}

	return m.Ctr.WithExec(args)
}

type WithSource struct {
	// +private
	Bats *Bats
}

func (m *WithSource) Container() *Container {
	return m.Bats.Ctr
}

// Run bats tests.
func (m *WithSource) Run(
	// Arguments to pass to bats.
	args []string,
) *Container {
	return m.Bats.Ctr.WithExec(args)
}
