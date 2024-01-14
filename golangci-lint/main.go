package main

import (
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "golangci/golangci-lint"

type GolangciLint struct {
	// +private
	Go *Go
}

func New(
	// Version (image tag) to use from the official image repository as a golangci-lint binary source.
	// +optional
	version string,

	// Custom image reference in "repository:tag" format to use as a golangci-lint binary source.
	// +optional
	image string,

	// Custom image reference in "repository:tag" format to use as a golangci-lint binary source.
	// +optional
	container *Container,

	// Disable mounting cache volumes.
	// +optional
	disableCache bool,

	// Linter cache volume to mount at ~/.cache/golangci-lint.
	// +optional
	linterCache *CacheVolume,

	// Version (image tag) to use from the official image repository as a Go base container.
	// +optional
	goVersion string,

	// Custom image reference in "repository:tag" format to use as a Go base container.
	// +optional
	goImage string,

	// Custom container to use as a Go base container.
	// +optional
	goContainer *Container,

	// Disable mounting Go cache volumes.
	// +optional
	disableGoCache bool,
) *GolangciLint {
	var golangciLint *Container

	if version != "" {
		golangciLint = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	} else if image != "" {
		golangciLint = dag.Container().From(image)
	} else if container != nil {
		golangciLint = container
	} else {
		golangciLint = dag.Container().From(defaultImageRepository)
	}

	ctr := dag.Go(GoOpts{
		Version:      goVersion,
		Image:        goImage,
		Container:    goContainer,
		DisableCache: disableCache || disableGoCache,
	}).Container().
		WithFile("/usr/local/bin/golangci-lint", golangciLint.File("/usr/bin/golangci-lint")).
		WithoutEnvVariable("GOLANGCI_LINT_CACHE"). // Make sure golangci-lint cache location is not overridden
		With(func(c *Container) *Container {
			if !disableCache {
				return c.WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint"))
			}

			return c
		})

	return &GolangciLint{dag.Go(GoOpts{Container: ctr})}
}

func (m *GolangciLint) Container() *Container {
	return m.Go.Container()
}

// Mount a cache volume for golangci-lint cache.
func (m *GolangciLint) WithLinterCache(
	cache *CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	// +optional
	source *Directory,

	// Sharing mode of the cache volume.
	// +optional
	sharing CacheSharingMode,
) *GolangciLint {
	return &GolangciLint{
		dag.Go(GoOpts{
			Container: m.Go.Container().WithMountedCache("/root/.cache/golangci-lint", cache, ContainerWithMountedCacheOpts{
				Source:  source,
				Sharing: sharing,
			}),
		}),
	}
}

// Mount a cache volume for Go module cache.
func (m *GolangciLint) WithModuleCache(
	cache *CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	// +optional
	source *Directory,

	// Sharing mode of the cache volume.
	// +optional
	sharing CacheSharingMode,
) *GolangciLint {
	return &GolangciLint{m.Go.WithModuleCache(cache, GoWithModuleCacheOpts{
		Source:  source,
		Sharing: string(sharing),
	})}
}

// Mount a cache volume for Go build cache.
func (m *GolangciLint) WithBuildCache(
	cache *CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	// +optional
	source *Directory,

	// Sharing mode of the cache volume.
	// +optional
	sharing CacheSharingMode,
) *GolangciLint {
	return &GolangciLint{m.Go.WithBuildCache(cache, GoWithBuildCacheOpts{
		Source:  source,
		Sharing: string(sharing),
	})}
}

func (m *GolangciLint) Run(
	source *Directory,

	// Read custom configuration file.
	// +optional
	config *File,

	// Timeout for total work
	// +optional
	timeout string,

	// Verbose output
	// +optional
	verbose bool,

	// Additional arguments to pass to the command.
	// +optional
	rawArgs []string,
) *Container {
	args := []string{"golangci-lint", "run"}

	if config != nil {
		args = append(args, "--config", "/work/config")
	}

	if timeout != "" {
		args = append(args, "--timeout", timeout)
	}

	if verbose {
		args = append(args, "--verbose")
	}

	if len(rawArgs) > 0 {
		args = append(args, rawArgs...)
	}

	return m.Go.Container().
		WithWorkdir("/work/src").
		WithMountedDirectory("/work/src", source).
		With(func(c *Container) *Container {
			if config != nil {
				c = c.WithMountedFile("/work/config", config)
			}

			return c
		}).
		WithExec(args)
}
