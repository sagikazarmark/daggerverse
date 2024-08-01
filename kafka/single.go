package main

import (
	"dagger/kafka/internal/dagger"
	"fmt"
)

// Launch a single-node Kafka cluster.
func (m *Kafka) SingleNode(
	// Name of the service.
	//
	// +default="kafka"
	// +optional
	serviceName string,
) *SingleNode {
	if serviceName == "" {
		serviceName = "kafka"
	}

	return &SingleNode{
		ServiceName: serviceName,
		Ctr:         m.Container,
	}
}

// A single-node Kafka cluster.
type SingleNode struct {
	ServiceName string

	// +private
	Ctr *dagger.Container
}

// Set an environment variable.
func (m *SingleNode) WithEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,

	// The value of the environment variable (e.g., "localhost").
	value string,

	// Replace `${VAR}` or $VAR in the value according to the current environment
	// variables defined in the container (e.g., "/opt/bin:$PATH").
	//
	// +optional
	expand bool,
) *SingleNode {
	m.Ctr = m.Ctr.WithEnvVariable(name, value, dagger.ContainerWithEnvVariableOpts{
		Expand: expand,
	})

	return m
}

func (m *SingleNode) Container() *dagger.Container {
	return m.Ctr.
		// KRaft settings
		WithEnvVariable("KAFKA_CFG_NODE_ID", "0").
		WithEnvVariable("KAFKA_CFG_PROCESS_ROLES", "controller,broker").
		WithEnvVariable("KAFKA_CFG_CONTROLLER_QUORUM_VOTERS", "0@127.0.0.1:9093").

		// Listeners
		WithEnvVariable("KAFKA_CFG_LISTENERS", "PLAINTEXT://:9092,CONTROLLER://:9093").
		WithEnvVariable("KAFKA_CFG_ADVERTISED_LISTENERS", fmt.Sprintf("PLAINTEXT://%s:9092", m.ServiceName)).
		WithEnvVariable("KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP", "CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT").
		WithEnvVariable("KAFKA_CFG_CONTROLLER_LISTENER_NAMES", "CONTROLLER").
		WithEnvVariable("KAFKA_CFG_INTER_BROKER_LISTENER_NAME", "PLAINTEXT")
}

// Launch the Kafka service.
func (m *SingleNode) Service() *dagger.Service {
	return m.Container().
		WithExposedPort(9092).
		WithExposedPort(9093).
		AsService()
}
