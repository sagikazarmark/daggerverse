package main

import (
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "bats/bats"

type Bats struct{}

// Specify which version (image tag) of Bats to use from the official image repository on Docker Hub.
func (m *Bats) FromVersion(version string) *Base {
	return &Base{dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))}
}

// Specify a custom image reference in "repository:tag" format.
func (m *Bats) FromImage(image string) *Base {
	return &Base{dag.Container().From(image)}
}

// Specify a custom container.
func (m *Bats) FromContainer(ctr *Container) *Base {
	return &Base{ctr}
}

func defaultContainer() *Base {
	return &Base{dag.Container().From(defaultImageRepository)}
}

// Mount a source directory.
func (m *Bats) WithSource(src *Directory) *BaseWithSource {
	return defaultContainer().WithSource(src)
}

func (m *Bats) Run(args []string, version Optional[string], image Optional[string], container Optional[*Container], source Optional[*Directory]) *Container {
	var base *Base

	if v, ok := version.Get(); ok {
		base = m.FromVersion(v)
	} else if i, ok := image.Get(); ok {
		base = m.FromImage(i)
	} else if c, ok := container.Get(); ok {
		base = m.FromContainer(c)
	} else {
		base = defaultContainer()
	}

	return base.Run(args, source)
}

// Return the default container.
func (m *Bats) Container() *Container {
	return defaultContainer().Container()
}

type Base struct {
	Ctr *Container
}

// Return the underlying container.
func (m *Base) Container() *Container {
	return m.Ctr
}

// Mount a source directory.
func (m *Base) WithSource(src *Directory) *BaseWithSource {
	const workdir = "/src"

	return &BaseWithSource{
		&Base{
			m.Ctr.
				WithWorkdir(workdir).
				WithMountedDirectory(workdir, src),
		},
	}
}

func (m *Base) Run(args []string, source Optional[*Directory]) *Container {
	ctr := m.Ctr

	if src, ok := source.Get(); ok {
		ctr = m.WithSource(src).Ctr
	}

	return ctr.WithExec(args)
}

type BaseWithSource struct {
	*Base
}

func (m *BaseWithSource) Run(args []string) *Container {
	return m.Ctr.WithExec(args)
}
