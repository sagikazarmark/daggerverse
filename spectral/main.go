package main

import (
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "stoplight/spectral"

type Spectral struct{}

// Specify which version (image tag) of Spectral to use from the official image repository on Docker Hub.
func (m *Spectral) FromVersion(version string) *Base {
	return &Base{dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))}
}

// Specify a custom image reference in "repository:tag" format.
func (m *Spectral) FromImage(image string) *Base {
	return &Base{dag.Container().From(image)}
}

// Specify a custom container.
func (m *Spectral) FromContainer(ctr *Container) *Base {
	return &Base{ctr}
}

func defaultContainer() *Base {
	return &Base{dag.Container().From(defaultImageRepository)}
}

// Mount a source directory.
func (m *Spectral) WithSource(src *Directory) *BaseWithSource {
	return defaultContainer().WithSource(src)
}

func (m *Spectral) Lint(document string, version Optional[string], image Optional[string], container Optional[*Container], source Optional[*Directory]) *Container {
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

	return base.Lint(document, source)
}

// Return the default container.
func (m *Spectral) Container() *Container {
	return defaultContainer().Container()
}

type Base struct {
	Ctr *Container
}

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

func (m *Base) Lint(document string, source Optional[*Directory]) *Container {
	ctr := m.Ctr

	if src, ok := source.Get(); ok {
		ctr = m.WithSource(src).Ctr
	}

	return lint(ctr, document)
}

type BaseWithSource struct {
	*Base
}

// example usage: "dagger call with-source --src . lint --document openapi.yaml"
func (m *BaseWithSource) Lint(document string) *Container {
	return lint(m.Ctr, document)
}

func lint(ctr *Container, document string) *Container {
	args := []string{"lint"}
	args = append(args, document)

	return ctr.WithExec(args)
}
