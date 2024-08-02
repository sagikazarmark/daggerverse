package main

// Launch a single-node Kafka cluster.
func (m *Kafka) SingleNode(
	// Name of the service.
	//
	// +default="kafka"
	// +optional
	serviceName string,
) *Node {
	if serviceName == "" {
		serviceName = "kafka"
	}

	return &Node{
		ServiceName: serviceName,
		NodeID:      0,
		Ctr:         m.Container,
	}
}
