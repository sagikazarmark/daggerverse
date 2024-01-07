package main

import (
	"fmt"
	"strings"

	"github.com/containerd/containerd/platforms"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "golang"

type Go struct {
	Ctr *Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	version Optional[string],

	// Custom image reference in "repository:tag" format to use as a base container.
	image Optional[string],

	// Custom container to use as a base container.
	container Optional[*Container],

	// Disable mounting cache volumes.
	disableCache Optional[bool],

	// Module cache volume to mount at /go/pkg/mod.
	modCache Optional[*CacheVolume],

	// Build cache volume to mount at ~/.cache/go-build.
	buildCache Optional[*CacheVolume],
) *Go {
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

	if !disableCache.GetOr(false) {
		ctr = ctr.
			WithMountedCache("/go/pkg/mod", modCache.GetOr(dag.CacheVolume("go-mod"))).
			WithMountedCache("/root/.cache/go-build", buildCache.GetOr(dag.CacheVolume("go-build")))
	}

	return &Go{
		Ctr: ctr,
	}
}

// DEPRECATED: Specify which version (image tag) of Go to use from the official image repository on Docker Hub.
func (m *Go) FromVersion(version string) *Go {
	return New(Opt(version), OptEmpty[string](), OptEmpty[*Container](), OptEmpty[bool](), OptEmpty[*CacheVolume](), OptEmpty[*CacheVolume]())
}

// DEPRECATED: Specify a custom image reference in "repository:tag" format.
func (m *Go) FromImage(image string) *Go {
	return New(OptEmpty[string](), Opt(image), OptEmpty[*Container](), OptEmpty[bool](), OptEmpty[*CacheVolume](), OptEmpty[*CacheVolume]())
}

// DEPRECATED: Specify a custom container.
func (m *Go) FromContainer(ctr *Container) *Go {
	return New(OptEmpty[string](), OptEmpty[string](), Opt(ctr), OptEmpty[bool](), OptEmpty[*CacheVolume](), OptEmpty[*CacheVolume]())
}

// Set an environment variable.
func (m *Go) WithEnvVariable(name string, value string, expand Optional[bool]) *Go {
	return &Go{
		m.Ctr.WithEnvVariable(name, value, ContainerWithEnvVariableOpts{
			Expand: expand.GetOr(false),
		}),
	}
}

// Set GOOS, GOARCH and GOARM environment variables.
func (m *Go) WithPlatform(platform Platform) *Go {
	if platform == "" {
		return m
	}

	p := platforms.MustParse(string(platform))

	ctr := m.Ctr.
		WithEnvVariable("GOOS", p.OS).
		WithEnvVariable("GOARCH", p.Architecture)

	if p.Variant != "" {
		ctr = ctr.WithEnvVariable("GOARM", p.Variant)
	}

	return &Go{ctr}
}

// Set CGO_ENABLED environment variable to 1.
func (m *Go) WithCgoEnabled() *Go {
	return &Go{m.Ctr.WithEnvVariable("CGO_ENABLED", "1")}
}

// Set CGO_ENABLED environment variable to 0.
func (m *Go) WithCgoDisabled() *Go {
	return &Go{m.Ctr.WithEnvVariable("CGO_ENABLED", "0")}
}

// Mount a source directory.
func (m *Go) WithSource(src *Directory) *WithSource {
	const workdir = "/src"

	return &WithSource{
		&Go{
			m.Ctr.
				WithWorkdir(workdir).
				WithMountedDirectory(workdir, src),
		},
	}
}

// Run a Go command.
func (m *Go) Exec(args []string, source Optional[*Directory], platform Optional[Platform]) *Container {
	if src, ok := source.Get(); ok {
		return m.WithSource(src).Exec(args, platform)
	}

	if p, ok := platform.Get(); ok {
		m = m.WithPlatform(p)
	}

	return m.Ctr.WithExec(args)
}

// Run "go build" command.
func (m *Go) Build(
	source *Directory,
	pkg Optional[string],
	tags Optional[[]string],
	trimpath Optional[bool],
	rawArgs Optional[[]string],
	platform Optional[Platform],
) *File {
	return m.WithSource(source).Build(pkg, tags, trimpath, rawArgs, platform)
}

// Return the base container.
func (m *Go) Container() *Container {
	return m.Ctr
}

type WithSource struct {
	Go *Go
}

// Set an environment variable.
func (m *WithSource) WithEnvVariable(name string, value string, expand Optional[bool]) *WithSource {
	return &WithSource{m.Go.WithEnvVariable(name, value, expand)}
}

// Set GOOS, GOARCH and GOARM environment variables.
func (m *WithSource) WithPlatform(platform Platform) *WithSource {
	return &WithSource{m.Go.WithPlatform(platform)}
}

// Set CGO_ENABLED environment variable to 1.
func (m *WithSource) WithCgoEnabled() *WithSource {
	return &WithSource{m.Go.WithCgoEnabled()}
}

// Set CGO_ENABLED environment variable to 0.
func (m *WithSource) WithCgoDisabled() *WithSource {
	return &WithSource{m.Go.WithCgoDisabled()}
}

func (m *WithSource) Exec(args []string, platform Optional[Platform]) *Container {
	if p, ok := platform.Get(); ok {
		m = m.WithPlatform(p)
	}

	return m.Go.Container().WithExec(args)
}

func (m *WithSource) Build(pkg Optional[string], tags Optional[[]string], trimpath Optional[bool], rawArgs Optional[[]string], platform Optional[Platform]) *File {
	args := []string{"go", "build", "-o", "/out/result"}

	if tags, ok := tags.Get(); ok && len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}

	if trimpath.GetOr(false) {
		args = append(args, "-trimpath")
	}

	if rawArgs, ok := rawArgs.Get(); ok {
		args = append(args, rawArgs...)
	}

	if pkg, ok := pkg.Get(); ok {
		args = append(args, pkg)
	}

	return m.Exec(args, platform).File("/out/result")
}
