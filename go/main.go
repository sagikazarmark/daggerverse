package main

import (
	"fmt"

	"github.com/containerd/containerd/platforms"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "golang"

type Go struct{}

// Specify which version (image tag) of Go to use from the official image repository on Docker Hub.
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

func defaultContainer() *Base {
	return &Base{wrapContainer(dag.Container().From(defaultImageRepository))}
}

func wrapContainer(c *Container) *Container {
	return c.
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod"))
}

// Set an environment variable.
func (m *Go) WithEnvVariable(name string, value string, expand Optional[bool]) *Base {
	return defaultContainer().WithEnvVariable(name, value, expand)
}

// Set GOOS, GOARCH and GOARM environment variables.
func (m *Go) WithPlatform(platform Platform) *Base {
	return defaultContainer().WithPlatform(platform)
}

// Mount a source directory. The container will use the latest official Go image.
func (m *Go) WithSource(src *Directory) *BaseWithSource {
	return defaultContainer().WithSource(src)
}

// Run a Go command in a container (default: latest official Go image with no mounted source).
func (m *Go) Exec(args []string, version Optional[string], image Optional[string], container Optional[*Container], source Optional[*Directory], platform Optional[Platform]) *Container {
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

	if p, ok := platform.Get(); ok {
		base = base.WithPlatform(p)
	}

	return base.Exec(args, source)
}

// Return the default container.
func (m *Go) Container() *Container {
	return defaultContainer().Container()
}

type Base struct {
	Ctr *Container
}

// Return the current container.
func (m *Base) Container() *Container {
	return m.Ctr
}

// Set an environment variable.
func (m *Base) WithEnvVariable(name string, value string, expand Optional[bool]) *Base {
	return &Base{
		m.Ctr.WithEnvVariable(name, value, ContainerWithEnvVariableOpts{
			Expand: expand.GetOr(false),
		}),
	}
}

// Set GOOS, GOARCH and GOARM environment variables.
func (m *Base) WithPlatform(platform Platform) *Base {
	p := platforms.MustParse(string(platform))

	ctr := m.Ctr.
		WithEnvVariable("GOOS", p.OS).
		WithEnvVariable("GOARCH", p.Architecture)

	if p.Variant != "" {
		ctr = ctr.WithEnvVariable("GOARM", p.Variant)
	}

	return &Base{
		ctr,
	}
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

func (m *Base) Exec(args []string, source Optional[*Directory]) *Container {
	ctr := m.Ctr

	if src, ok := source.Get(); ok {
		ctr = m.WithSource(src).Ctr
	}

	return ctr.WithExec(args)
}

type BaseWithSource struct {
	*Base
}

// Set an environment variable.
func (m *BaseWithSource) WithEnvVariable(name string, value string, expand Optional[bool]) *BaseWithSource {
	return &BaseWithSource{m.Base.WithEnvVariable(name, value, expand)}
}

func (m *BaseWithSource) Exec(args []string) *Container {
	return m.Ctr.WithExec(args)
}
