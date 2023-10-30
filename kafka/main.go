package main

import "fmt"

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "bitnami/kafka"

type Kafka struct{}

// Specify which version of Kafka to use.
func (m *Kafka) WithVersion(version string) *KafkaContainer {
	return &KafkaContainer{dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))}
}

// Specify a custom image reference in "repository:tag" format.
//
// Note: this module expects an image compatible with bitnami/kafka.
func (m *Kafka) WithImageRef(ref string) *KafkaContainer {
	return &KafkaContainer{dag.Container().From(ref)}
}

// Specify a custom container.
//
// Note: this module expects an image compatible with bitnami/kafka.
func (m *Kafka) WithContainer(ctr *Container) *KafkaContainer {
	return &KafkaContainer{ctr}
}

// Return the default container.
func (m *Kafka) Container() *Container {
	return defaultContainer().Container()
}

// Launch a Kafka service using the default container.
func (m *Kafka) Service() *Service {
	return defaultContainer().Service()
}

func defaultContainer() *KafkaContainer {
	return &KafkaContainer{dag.Container().From(defaultImageRepository)}
}

type KafkaContainer struct {
	Ctr *Container
}

// Return the underlying container.
func (m *KafkaContainer) Container() *Container {
	return m.Ctr
}

// Launch a Kafka service container.
func (m *KafkaContainer) Service() *Service {
	return m.Ctr.
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
		WithExposedPort(9092).
		AsService()
}
