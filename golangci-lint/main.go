// Fast linters runner for Go.
package main

import (
	"dagger/golangci-lint/internal/dagger"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "golangci/golangci-lint"

type GolangciLint struct {
	Binary    *dagger.File
	Container *dagger.Container
}

const cachePath = "/var/cache/golangci-lint"

func New(
	// Version (image tag) to use from the official image repository as a golangci-lint binary source.
	//
	// +optional
	version string,

	// Custom container to use as a golangci-lint binary source.
	//
	// +optional
	container *dagger.Container,

	// golangci-lint binary.
	//
	// +optional
	binary *dagger.File,

	// Disable mounting default cache volumes.
	//
	// +optional
	disableCache bool,

	// Linter cache volume to mount (takes precedence over disableCache).
	//
	// +optional
	cache *dagger.CacheVolume,

	// Version (image tag) to use from the official image repository as a Go base container.
	//
	// +optional
	goVersion string,

	// Custom container to use as a Go base container.
	//
	// +optional
	goContainer *dagger.Container,

	// Disable mounting default Go cache volumes.
	//
	// +optional
	disableGoCache bool,
) *GolangciLint {
	if binary == nil {
		if container == nil {
			if version == "" {
				version = "latest"
			}

			container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
		}

		binary = container.File("/usr/bin/golangci-lint")
	}

	if !disableCache && cache == nil {
		cache = dag.CacheVolume("golangci-lint")
	}

	container = dag.Go(dagger.GoOpts{
		Version:      goVersion,
		Container:    goContainer,
		DisableCache: disableCache || disableGoCache,
	}).Container().
		WithEnvVariable("GOLANGCI_LINT_CACHE", cachePath). // Make sure golangci-lint cache location is not overridden
		WithFile("/usr/local/bin/golangci-lint", binary)

	m := &GolangciLint{
		Binary:    binary,
		Container: container,
	}

	if cache != nil {
		m = m.WithCache(cache, nil, "")
	}

	return m
}

// Mount a cache volume for golangci-lint cache.
func (m *GolangciLint) WithCache(
	cache *dagger.CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	//
	// +optional
	source *dagger.Directory,

	// Sharing mode of the cache volume.
	//
	// +optional
	sharing dagger.CacheSharingMode,
) *GolangciLint {
	m.Container = m.Container.WithMountedCache(cachePath, cache, dagger.ContainerWithMountedCacheOpts{
		Source:  source,
		Sharing: sharing,
	})

	return m
}

// Run the linters.
func (m *GolangciLint) Run(
	source *dagger.Directory,

	// Read custom configuration file.
	//
	// +optional
	config *dagger.File,

	// Timeout for total work.
	//
	// +optional
	timeout string,

	// Verbose output.
	//
	// +optional
	verbose bool,

	// Additional arguments to pass to the command.
	//
	// +optional
	rawArgs []string,
) *dagger.Container {
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

	return m.Container.
		WithWorkdir("/work/src").
		WithMountedDirectory(".", source).
		With(func(c *dagger.Container) *dagger.Container {
			if config != nil {
				c = c.WithMountedFile("/work/config", config)
			}

			return c
		}).
		WithExec(args)
}

// Format Go source files.
func (m *GolangciLint) Fmt(
	source *dagger.Directory,

	// Read custom configuration file.
	//
	// +optional
	config *dagger.File,

	// Additional arguments to pass to the command.
	//
	// +optional
	rawArgs []string,
) *dagger.Directory {
	args := []string{"golangci-lint", "fmt"}

	if config != nil {
		args = append(args, "--config", "/work/config")
	}

	if len(rawArgs) > 0 {
		args = append(args, rawArgs...)
	}

	return m.Container.
		WithWorkdir("/work/src").
		WithMountedDirectory(".", source).
		With(func(c *dagger.Container) *dagger.Container {
			if config != nil {
				c = c.WithMountedFile("/work/config", config)
			}

			return c
		}).
		WithExec(args).
		Directory(".")
}
