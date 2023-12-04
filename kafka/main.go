package main

import "fmt"

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "bitnami/kafka"

type Kafka struct{}

// Specify which version (image tag) of Kafka to use from the official image repository on Docker Hub.
func (m *Kafka) FromVersion(version string) *Base {
	return &Base{wrapContainer(dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version)))}
}

// Specify a custom image reference in "repository:tag" format.
//
// Note: this module expects an image compatible with bitnami/kafka.
func (m *Kafka) FromImage(ref string) *Base {
	return &Base{wrapContainer(dag.Container().From(ref))}
}

// Specify a custom container.
//
// Note: this module expects an image compatible with bitnami/kafka.
func (m *Kafka) FromContainer(ctr *Container) *Base {
	return &Base{wrapContainer(ctr)}
}

func defaultContainer() *Base {
	return &Base{wrapContainer(dag.Container().From(defaultImageRepository))}
}

func wrapContainer(c *Container) *Container {
	return c.
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
}

// Set an environment variable.
func (m *Kafka) WithEnvVariable(name string, value string, expand Optional[bool]) *Base {
	return defaultContainer().WithEnvVariable(name, value, expand)
}

// Return the default container.
func (m *Kafka) Container() *Container {
	return defaultContainer().Container()
}

// Launch a Kafka service using the default container.
func (m *Kafka) Service() *Service {
	return defaultContainer().Service()
}

type Base struct {
	Ctr *Container
}

// Return the underlying container.
func (m *Base) Container() *Container {
	return m.Ctr
}

// Set an environment variable.
func (m *Base) WithEnvVariable(name string, value string, expand Optional[bool]) *Base {
	return &Base{
		m.Ctr.WithEnvVariable(name, value, ContainerWithEnvVariableOpts{
			Expand: expand.GetOr(false),
		}),
	}
}

// Launch a Kafka service container.
func (m *Base) Service() *Service {
	return m.Ctr.AsService()
}
