package main

import (
	"context"
	"fmt"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.SingleNode_Connect)
	p.Go(m.Cluster_Connect)

	return p.Wait()
}

func (m *Tests) SingleNode_Connect(ctx context.Context) error {
	cluster := dag.Kafka().SingleNode()
	serviceName, err := cluster.ServiceName(ctx)
	if err != nil {
		return err
	}

	_, err = dag.Kafka().Container().
		WithServiceBinding(serviceName, cluster.Service()).
		WithExec([]string{"kafka-topics.sh", "--list", "--bootstrap-server", fmt.Sprintf("%s:9092", serviceName)}).
		Sync(ctx)

	return err
}

func (m *Tests) Cluster_Connect(ctx context.Context) error {
	cluster := dag.Kafka().Cluster()

	nodes, err := cluster.Nodes(ctx)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		serviceName, err := node.ServiceName(ctx)
		if err != nil {
			return err
		}

		_, err = dag.Kafka().Container().
			WithServiceBinding(serviceName, node.Service()).
			WithExec([]string{"kafka-topics.sh", "--list", "--bootstrap-server", fmt.Sprintf("%s:9092", serviceName)}).
			Sync(ctx)

		if err != nil {
			return err
		}
	}

	return nil
}
