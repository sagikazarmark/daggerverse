// The Registry is a stateless, highly scalable server side application that stores and lets you distribute container images and other content.

package main

import (
	"context"
	"dagger/registry/internal/dagger"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "registry"

type Registry struct {
	// +private
	Container *dagger.Container
}

func New(
	ctx context.Context,

	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	// +default="2.8"
	version string,

	// Custom container to use as a base container. Takes precedence over version.
	//
	// +optional
	container *dagger.Container,

	// Port to expose the registry on.
	//
	// +optional
	// +default=5000
	port int,

	// Mount a volume to persist data between runs.
	//
	// +optional
	dataVolume *dagger.CacheVolume,
) (*Registry, error) {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(defaultImageRepository + ":" + version)
	}

	if port == 0 {
		port = 5000
	}

	container = container.
		WithExposedPort(port).
		WithEnvVariable("REGISTRY_HTTP_ADDR", fmt.Sprintf("0.0.0.0:%d", port)).
		With(func(c *dagger.Container) *dagger.Container {
			if dataVolume != nil {
				c = c.
					WithEnvVariable("REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY", "/var/lib/registry").
					WithMountedCache("/var/lib/registry", dataVolume)
			}

			return c
		})

	return &Registry{
		Container: container,
	}, nil
}

func (m *Registry) Service() *dagger.Service {
	return m.Container.AsService()
}
