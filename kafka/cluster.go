package main

import (
	"dagger/kafka/internal/dagger"
	"fmt"
)

// Launch a Kafka cluster.
func (m *Kafka) Cluster(
	// Name of the service.
	//
	// +default="kafka"
	// +optional
	serviceNamePrefix string,

	// Number of nodes in the cluster.
	//
	// +default=3
	// +optional
	nodeCount int,
) *Cluster {
	if serviceNamePrefix == "" {
		serviceNamePrefix = "kafka"
	}

	if nodeCount < 1 {
		nodeCount = 3
	}

	return &Cluster{
		ServiceNamePrefix: serviceNamePrefix,
		NodeCount:         nodeCount,
		Container:         m.Container,
	}
}

// A Kafka cluster.
type Cluster struct {
	// +private
	ServiceNamePrefix string

	// +private
	NodeCount int

	// +private
	Container *dagger.Container
}

// Set an environment variable.
func (m *Cluster) WithEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,

	// The value of the environment variable (e.g., "localhost").
	value string,

	// Replace `${VAR}` or $VAR in the value according to the current environment
	// variables defined in the container (e.g., "/opt/bin:$PATH").
	//
	// +optional
	expand bool,
) *Cluster {
	m.Container = m.Container.WithEnvVariable(name, value, dagger.ContainerWithEnvVariableOpts{
		Expand: expand,
	})

	return m
}

// Return the Kafka nodes.
func (m *Cluster) Nodes() []*Node {
	var controllerService *dagger.Service

	nodes := make([]*Node, 0, m.NodeCount)

	for i := 0; i < m.NodeCount; i++ {
		node := &Node{
			ServiceName: fmt.Sprintf("%s%d", m.ServiceNamePrefix, i),
			NodeID:      i,
			Ctr:         m.Container,
		}

		if i == 0 {
			controllerService = node.Service()
		} else {
			node.Ctr = node.Ctr.WithServiceBinding("controller", controllerService)
		}

		nodes = append(nodes, node)
	}

	return nodes
}

// A Kafka node.
type Node struct {
	ServiceName string

	// +private
	NodeID int

	// +private
	Ctr *dagger.Container
}

// Set an environment variable.
func (m *Node) WithEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,

	// The value of the environment variable (e.g., "localhost").
	value string,

	// Replace `${VAR}` or $VAR in the value according to the current environment
	// variables defined in the container (e.g., "/opt/bin:$PATH").
	//
	// +optional
	expand bool,
) *Node {
	m.Ctr = m.Ctr.WithEnvVariable(name, value, dagger.ContainerWithEnvVariableOpts{
		Expand: expand,
	})

	return m
}

func (m *Node) Container() *dagger.Container {
	return m.Ctr.
		// KRaft settings
		WithEnvVariable("KAFKA_CFG_NODE_ID", fmt.Sprintf("%d", m.NodeID)).
		WithEnvVariable("KAFKA_KRAFT_CLUSTER_ID", "ciWo7IWazngRchmPES6q5A=="). // TODO: generate this?
		With(func(c *dagger.Container) *dagger.Container {
			// Node 0 acts as the only controller for now, because it's not possible to have circular references between services.
			if m.NodeID == 0 {
				c = c.
					WithEnvVariable("KAFKA_CFG_PROCESS_ROLES", "controller,broker").
					WithEnvVariable("KAFKA_CFG_CONTROLLER_QUORUM_VOTERS", "0@127.0.0.1:9093")
			} else {
				c = c.
					WithEnvVariable("KAFKA_CFG_PROCESS_ROLES", "broker").
					WithEnvVariable("KAFKA_CFG_CONTROLLER_QUORUM_VOTERS", "0@controller:9093")
			}

			return c
		}).

		// Listeners
		WithEnvVariable("KAFKA_CFG_ADVERTISED_LISTENERS", fmt.Sprintf("PLAINTEXT://%s:9092", m.ServiceName)).
		WithEnvVariable("KAFKA_CFG_INTER_BROKER_LISTENER_NAME", "PLAINTEXT").
		WithEnvVariable("KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP", "CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT"). // Why is this needed for broker only nodes?
		WithEnvVariable("KAFKA_CFG_CONTROLLER_LISTENER_NAMES", "CONTROLLER").
		With(func(c *dagger.Container) *dagger.Container {
			// Node 0 acts as the only controller for now, because it's not possible to have circular references between services.
			if m.NodeID == 0 {
				c = c.
					WithEnvVariable("KAFKA_CFG_LISTENERS", "PLAINTEXT://:9092,CONTROLLER://:9093")
				// WithEnvVariable("KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP", "CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT").
				// WithEnvVariable("KAFKA_CFG_CONTROLLER_LISTENER_NAMES", "CONTROLLER")
			} else {
				c = c.
					WithEnvVariable("KAFKA_CFG_LISTENERS", "PLAINTEXT://:9092")
				// WithEnvVariable("KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP", "PLAINTEXT:PLAINTEXT")
			}

			return c
		})
}

// Launch the Kafka service.
func (m *Node) Service() *dagger.Service {
	return m.Container().
		WithExposedPort(9092).
		With(func(c *dagger.Container) *dagger.Container {
			// Node 0 acts as the only controller for now, because it's not possible to have circular references between services.
			if m.NodeID == 0 {
				c = c.WithExposedPort(9093)
			}

			return c
		}).
		AsService()
}
