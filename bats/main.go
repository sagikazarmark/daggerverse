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
	version Optional[string],

	// Custom image reference in "repository:tag" format to use as a base container.
	image Optional[string],

	// Custom container to use as a base container.
	container Optional[*Container],
) *Bats {
	var ctr *Container

	if v, ok := version.Get(); ok {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, v))
	} else if i, ok := image.Get(); ok {
		ctr = dag.Container().From(i)
	} else if c, ok := container.Get(); ok {
		ctr = c
	} else {
		ctr = dag.Container().From(defaultImageRepository)
	}

	return &Bats{
		Ctr: ctr,
	}
}

func (m *Bats) Container() *Container {
	return m.Ctr
}

// Mount a source directory.
func (m *Bats) WithSource(src *Directory) *WithSource {
	const workdir = "/src"

	return &WithSource{
		&Bats{
			m.Ctr.
				WithWorkdir(workdir).
				WithMountedDirectory(workdir, src),
		},
	}
}

func (m *Bats) Run(args []string, source Optional[*Directory]) *Container {
	if src, ok := source.Get(); ok {
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

func (m *WithSource) Run(args []string) *Container {
	return m.Bats.Ctr.WithExec(args)
}
