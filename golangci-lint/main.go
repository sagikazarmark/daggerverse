package main

import (
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "golangci/golangci-lint"

type GolangciLint struct {
	// +private
	Ctr *Container
}

func New(
	// Version (image tag) to use from the official image repository as a golangci-lint binary source.
	version Optional[string],

	// Custom image reference in "repository:tag" format to use as a golangci-lint binary source.
	image Optional[string],

	// Custom container to use as a golangci-lint binary source.
	container Optional[*Container],

	// Disable mounting cache volumes.
	disableCache Optional[bool],

	// Lint cache volume to mount at ~/.cache/golangci-lint.
	lintCache Optional[*CacheVolume],

	// Version (image tag) to use from the official image repository as a Go base container.
	goVersion Optional[string],

	// Custom image reference in "repository:tag" format to use as a Go base container.
	goImage Optional[string],

	// Custom container to use as a Go base container.
	goContainer Optional[*Container],

	// Disable mounting Go cache volumes.
	disableGoCache Optional[bool],

	// Module cache volume to mount at /go/pkg/mod.
	goModCache Optional[*CacheVolume],

	// Build cache volume to mount at ~/.cache/go-build.
	goBuildCache Optional[*CacheVolume],
) *GolangciLint {
	var golangciLint *Container

	if v, ok := version.Get(); ok {
		golangciLint = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, v))
	} else if i, ok := image.Get(); ok {
		golangciLint = dag.Container().From(i)
	} else if c, ok := container.Get(); ok {
		golangciLint = c
	} else {
		golangciLint = dag.Container().From(defaultImageRepository)
	}

	goModule := dag.Go(GoOpts{
		Version:      goVersion.GetOr(""),
		Image:        goImage.GetOr(""),
		Container:    goContainer.GetOr(nil),
		DisableCache: disableCache.GetOr(false) || disableGoCache.GetOr(false),
		ModCache:     goModCache.GetOr(nil),
		BuildCache:   goBuildCache.GetOr(nil),
	})

	ctr := goModule.Container().
		WithoutEnvVariable("GOLANGCI_LINT_CACHE"). // Make sure golangci-lint cache location is not overridden
		With(func(c *Container) *Container {
			if !disableCache.GetOr(false) {
				return c.WithMountedCache("/root/.cache/golangci-lint", lintCache.GetOr(dag.CacheVolume("golangci-lint")))
			}

			return c
		}).
		WithFile("/usr/local/bin/golangci-lint", golangciLint.File("/usr/bin/golangci-lint"))

	return &GolangciLint{
		Ctr: ctr,
	}
}

func (m *GolangciLint) Container() *Container {
	return m.Ctr
}

func (m *GolangciLint) Run(
	source *Directory,

	// Read custom configuration file.
	config Optional[*File],

	// Timeout for total work
	timeout Optional[string],

	// Verbose output
	verbose Optional[bool],

	// Additional arguments to pass to the command.
	rawArgs Optional[[]string],
) *Container {
	args := []string{"golangci-lint", "run"}

	if _, ok := config.Get(); ok {
		args = append(args, "--config", "/config")
	}

	if t, ok := timeout.Get(); ok {
		args = append(args, "--timeout", t)
	}

	if verbose.GetOr(false) {
		args = append(args, "--verbose")
	}

	if a, ok := rawArgs.Get(); ok {
		args = append(args, a...)
	}

	return m.Ctr.
		WithWorkdir("/src").
		WithMountedDirectory("/src", source).
		With(func(c *Container) *Container {
			if conf, ok := config.Get(); ok {
				c = c.WithMountedFile("/config", conf)
			}

			return c
		}).
		WithExec(args)
}
