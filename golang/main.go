package main

import "fmt"

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "golang"

type Golang struct{}

// Specify which version of Go to use.
func (m *Golang) WithVersion(version string) *GolangContainer {
	return &GolangContainer{dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))}
}

// Specify a custom image reference in "repository:tag" format.
func (m *Golang) WithImageRef(ref string) *GolangContainer {
	return &GolangContainer{dag.Container().From(ref)}
}

// Specify a custom container.
func (m *Golang) WithContainer(ctr *Container) *GolangContainer {
	return &GolangContainer{ctr}
}

// Mount a source directory.
func (m *Golang) WithSource(src *Directory) *GolangContainerWithSource {
	return defaultContainer().WithSource(src)
}

func (m *Golang) Exec(args []string) *Container {
	return defaultContainer().Exec(args)
}

// Return the default container.
func (m *Golang) Container() *Container {
	return defaultContainer().Container()
}

func defaultContainer() *GolangContainer {
	return &GolangContainer{dag.Container().From(defaultImageRepository)}
}

type GolangContainer struct {
	Ctr *Container
}

func (m *GolangContainer) Container() *Container {
	return m.Ctr
}

func (m *GolangContainer) Exec(args []string) *Container {
	return m.Ctr.WithExec(args)
}

// Mount a source directory.
func (m *GolangContainer) WithSource(src *Directory) *GolangContainerWithSource {
	const workdir = "/src"

	return &GolangContainerWithSource{
		&GolangContainer{
			m.Ctr.
				WithWorkdir(workdir).
				WithMountedDirectory(workdir, src),
		},
	}
}

type GolangContainerWithSource struct {
	*GolangContainer
}
