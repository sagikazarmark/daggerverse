package main

import "fmt"

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "golang"

type Go struct{}

// Specify which version of Go to use.
func (m *Go) WithVersion(version string) *GoContainer {
	return &GoContainer{dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))}
}

// Specify a custom image reference in "repository:tag" format.
func (m *Go) WithImageRef(ref string) *GoContainer {
	return &GoContainer{dag.Container().From(ref)}
}

// Specify a custom container.
func (m *Go) WithContainer(ctr *Container) *GoContainer {
	return &GoContainer{ctr}
}

// Mount a source directory.
func (m *Go) WithSource(src *Directory) *GoContainerWithSource {
	return defaultContainer().WithSource(src)
}

func (m *Go) Exec(args []string) *Container {
	return defaultContainer().Exec(args)
}

// Return the default container.
func (m *Go) Container() *Container {
	return defaultContainer().Container()
}

func defaultContainer() *GoContainer {
	return &GoContainer{dag.Container().From(defaultImageRepository)}
}

type GoContainer struct {
	Ctr *Container
}

func (m *GoContainer) Container() *Container {
	return m.Ctr
}

func (m *GoContainer) Exec(args []string) *Container {
	return m.Ctr.WithExec(args)
}

// Mount a source directory.
func (m *GoContainer) WithSource(src *Directory) *GoContainerWithSource {
	const workdir = "/src"

	return &GoContainerWithSource{
		&GoContainer{
			m.Ctr.
				WithWorkdir(workdir).
				WithMountedDirectory(workdir, src),
		},
	}
}

type GoContainerWithSource struct {
	*GoContainer
}
