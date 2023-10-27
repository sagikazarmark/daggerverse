package main

import (
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "bats/bats"

type Bats struct{}

// Specify which version of Bats to use.
func (m *Bats) WithVersion(version string) *BatsContainer {
	return &BatsContainer{dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))}
}

// Specify a custom image reference in "repository:tag" format.
func (m *Bats) WithImageRef(ref string) *BatsContainer {
	return &BatsContainer{dag.Container().From(ref)}
}

// Specify a custom container.
func (m *Bats) WithContainer(ctr *Container) *BatsContainer {
	return &BatsContainer{ctr}
}

// Mount a source directory.
func (m *Bats) WithSource(src *Directory) *BatsContainerWithSource {
	return defaultContainer().WithSource(src)
}

// Return the default container.
func (m *Bats) Container() *Container {
	return defaultContainer().Container()
}

func defaultContainer() *BatsContainer {
	return &BatsContainer{dag.Container().From(defaultImageRepository)}
}

type BatsContainer struct {
	Ctr *Container
}

// Return the underlying container.
func (m *BatsContainer) Container() *Container {
	return m.Ctr
}

// Mount a source directory.
func (m *BatsContainer) WithSource(src *Directory) *BatsContainerWithSource {
	const workdir = "/src"

	return &BatsContainerWithSource{
		&BatsContainer{
			m.Ctr.
				WithWorkdir(workdir).
				WithMountedDirectory(workdir, src),
		},
	}
}

type BatsContainerWithSource struct {
	*BatsContainer
}

// example usage: "dagger call with-source --src . lint --document openapi.yaml"
func (m *BatsContainerWithSource) Run(args []string) *Container {
	return m.Ctr.WithExec(args)
}
