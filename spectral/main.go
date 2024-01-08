package main

import (
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "stoplight/spectral"

type Spectral struct {
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
) *Spectral {
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

	return &Spectral{
		Ctr: ctr,
	}
}

func (m *Spectral) Container() *Container {
	return m.Ctr
}

// Mount a source directory.
func (m *Spectral) WithSource(src *Directory) *WithSource {
	const workdir = "/src"

	return &WithSource{
		&Spectral{
			m.Ctr.
				WithWorkdir(workdir).
				WithMountedDirectory(workdir, src),
		},
	}
}

func (m *Spectral) Lint(document string, source Optional[*Directory]) *Container {
	if src, ok := source.Get(); ok {
		return m.WithSource(src).Lint(document)
	}

	return lint(m.Ctr, document)
}

type WithSource struct {
	// +private
	Spectral *Spectral
}

// example usage: "dagger call with-source --src . lint --document openapi.yaml"
func (m *WithSource) Lint(document string) *Container {
	return lint(m.Spectral.Ctr, document)
}

func lint(ctr *Container, document string) *Container {
	args := []string{"lint"}
	args = append(args, document)

	return ctr.WithExec(args)
}
