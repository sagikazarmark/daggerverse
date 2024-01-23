package main

import "fmt"

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "bitnami/kafka"

type Kafka struct {
	// +private
	Ctr *Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	// +optional
	version string,

	// Custom image reference in "repository:tag" format to use as a base container.
	// +optional
	image string,

	// Custom container to use as a base container.
	// +optional
	container *Container,
) *Kafka {
	var ctr *Container

	if version != "" {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	} else if image != "" {
		ctr = dag.Container().From(image)
	} else if container != nil {
		ctr = container
	} else {
		ctr = dag.Container().From(defaultImageRepository)
	}

	ctr = ctr.
		// https://github.com/bitnami/charts/issues/22552#issuecomment-1905721850
		WithEnvVariable("KAFKA_CFG_MESSAGE_MAX_BYTES", "1048588").

		// KRaft settings
		WithEnvVariable("KAFKA_CFG_NODE_ID", "0").
		WithEnvVariable("KAFKA_CFG_PROCESS_ROLES", "controller,broker").
		WithEnvVariable("KAFKA_CFG_CONTROLLER_QUORUM_VOTERS", "0@127.0.0.1:9093").
		// Listeners
		WithEnvVariable("KAFKA_CFG_LISTENERS", "PLAINTEXT://:9092,CONTROLLER://:9093").
		WithEnvVariable("KAFKA_CFG_ADVERTISED_LISTENERS", "PLAINTEXT://kafka:9092").
		WithEnvVariable("KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP", "CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT").
		WithEnvVariable("KAFKA_CFG_CONTROLLER_LISTENER_NAMES", "CONTROLLER").
		WithEnvVariable("KAFKA_CFG_INTER_BROKER_LISTENER_NAME", "PLAINTEXT").
		WithExposedPort(9092)

	return &Kafka{ctr}
}

func (m *Kafka) Container() *Container {
	return m.Ctr
}

// Set an environment variable.
func (m *Kafka) WithEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,

	// The value of the environment variable (e.g., "localhost").
	value string,

	// Replace `${VAR}` or $VAR in the value according to the current environment
	// variables defined in the container (e.g., "/opt/bin:$PATH").
	// +optional
	expand bool,
) *Kafka {
	return &Kafka{
		m.Ctr.WithEnvVariable(name, value, ContainerWithEnvVariableOpts{
			Expand: expand,
		}),
	}
}

// Launch a Kafka service.
func (m *Kafka) Service() *Service {
	return m.Ctr.AsService()
}
