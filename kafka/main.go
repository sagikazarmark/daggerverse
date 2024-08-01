// Kafka service module for Dagger.
package main

import (
	"dagger/kafka/internal/dagger"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "bitnami/kafka"

type Kafka struct {
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
) *Kafka {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	return &Kafka{container}
}

// Set an environment variable.
func (m *Kafka) WithEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,

	// The value of the environment variable (e.g., "localhost").
	value string,

	// Replace `${VAR}` or $VAR in the value according to the current environment
	// variables defined in the container (e.g., "/opt/bin:$PATH").
	//
	// +optional
	expand bool,
) *Kafka {
	return &Kafka{
		m.Container.WithEnvVariable(name, value, dagger.ContainerWithEnvVariableOpts{
			Expand: expand,
		}),
	}
}
