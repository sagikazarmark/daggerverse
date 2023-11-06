package main

import "fmt"

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "golang"

type Go struct{}

// Specify which version of Go to use from the official Go image repository on Docker Hub.
func (m *Go) FromVersion(version string) *Base {
	return &Base{wrapContainer(dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version)))}
}

// Specify a custom image reference in "repository:tag" format.
func (m *Go) FromImage(image string) *Base {
	return &Base{wrapContainer(dag.Container().From(image))}
}

// Specify a custom container.
func (m *Go) FromContainer(ctr *Container) *Base {
	return &Base{wrapContainer(ctr)}
}

func wrapContainer(c *Container) *Container {
	return c.
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod"))
}

// Mount a source directory. The container will use the latest official Go image.
func (m *Go) WithSource(src *Directory) *BaseWithSource {
	return defaultContainer().WithSource(src)
}

// Run a Go command in a container.
// By default it falls back to using the latest official Go image with no mounted source.
// You can use --version, --image, --container and --source to customize the container.
func (m *Go) Exec(args []string, version Optional[string], image Optional[string], container Optional[*Container], source Optional[*Directory]) *Container {
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

	return base.Exec(args, source)
}

// Return the default container.
func (m *Go) Container() *Container {
	return defaultContainer().Container()
}

func defaultContainer() *Base {
	return &Base{wrapContainer(dag.Container().From(defaultImageRepository))}
}

type Base struct {
	Ctr *Container
}

func (m *Base) Container() *Container {
	return m.Ctr
}

func (m *Base) Exec(args []string, source Optional[*Directory]) *Container {
	ctr := m.Ctr

	if src, ok := source.Get(); ok {
		ctr = m.WithSource(src).Ctr
	}

	return ctr.WithExec(args)
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

type BaseWithSource struct {
	*Base
}

func (m *BaseWithSource) Exec(args []string) *Container {
	return m.Ctr.WithExec(args)
}
