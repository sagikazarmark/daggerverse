// Python programming language module.
package main

import (
	"dagger/python/internal/dagger"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "python"

type Python struct {
	Container *dagger.Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	version string,

	// Custom container to use as a base container.
	//
	// +optional
	container *dagger.Container,
) *Python {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	return &Python{container}
}

// Mount a cache volume for Pip cache.
func (m *Python) WithPipCache(
	cache *dagger.CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	//
	// +optional
	source *dagger.Directory,

	// Sharing mode of the cache volume.
	//
	// +optional
	sharing dagger.CacheSharingMode,
) *Python {
	return &Python{m.Container.WithMountedCache("/root/.cache/pip", cache, dagger.ContainerWithMountedCacheOpts{
		Source:  source,
		Sharing: sharing,
	})}
}

// Mount a source directory.
func (m *Python) WithSource(
	// Source directory to mount.
	source *dagger.Directory,
) *Python {
	const workdir = "/work"

	return &Python{
		m.Container.
			WithWorkdir(workdir).
			WithMountedDirectory(workdir, source),
	}
}
