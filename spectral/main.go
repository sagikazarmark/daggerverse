package main

import (
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "stoplight/spectral"

type Spectral struct{}

// example usage: "dagger call with-version --version 6.11.0"
func (m *Spectral) WithVersion(version string) *SpectralContainer {
	return &SpectralContainer{dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))}
}

// example usage: "dagger call with-image-ref --ref stoplight/spectral:6.11.0"
func (m *Spectral) WithImageRef(ref string) *SpectralContainer {
	return &SpectralContainer{dag.Container().From(ref)}
}

func (m *Spectral) WithContainer(ctr *Container) *SpectralContainer {
	return &SpectralContainer{ctr}
}

// example usage: "dagger call with-source --src ."
func (m *Spectral) WithSource(src *Directory) *SpectralContainerWithSource {
	return &SpectralContainerWithSource{
		&SpectralContainer{
			dag.Container().From(defaultImageRepository).
				WithWorkdir("/src").
				WithMountedDirectory("/src", src),
		},
	}
}

type SpectralContainer struct {
	Ctr *Container
}

func (m *SpectralContainer) Container() *Container {
	return m.Ctr
}

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
