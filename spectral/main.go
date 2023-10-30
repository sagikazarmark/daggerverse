package main

import (
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "stoplight/spectral"

type Spectral struct{}

// Specify which version of Spectral to use.
func (m *Spectral) WithVersion(version string) *SpectralContainer {
	return &SpectralContainer{dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))}
}

// Specify a custom image reference in "repository:tag" format.
func (m *Spectral) WithImageRef(ref string) *SpectralContainer {
	return &SpectralContainer{dag.Container().From(ref)}
}

// Specify a custom container.
func (m *Spectral) WithContainer(ctr *Container) *SpectralContainer {
	return &SpectralContainer{ctr}
}

// Mount a source directory.
func (m *Spectral) WithSource(src *Directory) *SpectralContainerWithSource {
	return defaultContainer().WithSource(src)
}

// Return the default container.
func (m *Spectral) Container() *Container {
	return defaultContainer().Container()
}

func defaultContainer() *SpectralContainer {
	return &SpectralContainer{dag.Container().From(defaultImageRepository)}
}

type SpectralContainer struct {
	Ctr *Container
}

func (m *SpectralContainer) Container() *Container {
	return m.Ctr
}

// Mount a source directory.
func (m *SpectralContainer) WithSource(src *Directory) *SpectralContainerWithSource {
	const workdir = "/src"

	return &SpectralContainerWithSource{
		&SpectralContainer{
			m.Ctr.
				WithWorkdir(workdir).
				WithMountedDirectory(workdir, src),
		},
	}
}

type SpectralContainerWithSource struct {
	*SpectralContainer
}

// example usage: "dagger call with-source --src . lint --document openapi.yaml"
func (m *SpectralContainerWithSource) Lint(document string) *Container {
	args := []string{"lint"}
	args = append(args, document)

	return m.Ctr.WithExec(args)
}
