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
	version Optional[string],

	// Custom image reference in "repository:tag" format to use as a base container.
	image Optional[string],

	// Custom container to use as a base container.
	container Optional[*Container],
) *Kafka {
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

	ctr = ctr.
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

	return &Kafka{
		Ctr: ctr,
	}
}

func (m *Kafka) Container() *Container {
	return m.Ctr
}

// Set an environment variable.
func (m *Kafka) WithEnvVariable(name string, value string, expand Optional[bool]) *Kafka {
	return &Kafka{
		m.Ctr.WithEnvVariable(name, value, ContainerWithEnvVariableOpts{
			Expand: expand.GetOr(false),
		}),
	}
}

// Launch a Kafka service.
func (m *Kafka) Service() *Service {
	return m.Ctr.AsService()
}
