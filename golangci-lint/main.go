// Fast linters runner for Go.
package main

import (
	"dagger/golangci-lint/internal/dagger"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "golangci/golangci-lint"

type GolangciLint struct {
	Binary *dagger.File

	// +private
	Cache CacheMount

	// +private
	Go *dagger.Go
}

type CacheMount struct {
	Volume  *dagger.CacheVolume
	Source  *dagger.Directory
	Sharing dagger.CacheSharingMode
}

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

	// Disable mounting Go cache volumes.
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

	return &GolangciLint{
		Binary: binary,
		Cache:  CacheMount{Volume: cache},
		Go: dag.Go(dagger.GoOpts{
			Version:      goVersion,
			Container:    goContainer,
			DisableCache: disableCache || disableGoCache,
		}),
	}
}

func (m *GolangciLint) Container() *dagger.Container {
	const cachePath = "/var/cache/golangci-lint"

	return m.Go.Container().
		WithEnvVariable("GOLANGCI_LINT_CACHE", cachePath). // Make sure golangci-lint cache location is not overridden
		WithFile("/usr/local/bin/golangci-lint", m.Binary).
		With(func(c *dagger.Container) *dagger.Container {
			if m.Cache.Volume != nil {
				return c.WithMountedCache(cachePath, m.Cache.Volume, dagger.ContainerWithMountedCacheOpts{
					Source:  m.Cache.Source,
					Sharing: m.Cache.Sharing,
				})
			}

			return c
		})
}

const linterCachePath = "/var/cache/golangci-lint"

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
	m.Cache = CacheMount{
		Volume:  cache,
		Source:  source,
		Sharing: sharing,
	}

	return m
}

// Mount a cache volume for Go module cache.
func (m *GolangciLint) WithModuleCache(
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
	m.Go = m.Go.WithModuleCache(cache, dagger.GoWithModuleCacheOpts{
		Source:  source,
		Sharing: sharing,
	})

	return m
}

// Mount a cache volume for Go build cache.
func (m *GolangciLint) WithBuildCache(
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
	m.Go = m.Go.WithBuildCache(cache, dagger.GoWithBuildCacheOpts{
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

	return m.Go.Container().
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

	return m.Go.Container().
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
