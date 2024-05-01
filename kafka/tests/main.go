package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	// p.Go(m.Default)

	return p.Wait()
}

func (m *Tests) Connect(ctx context.Context) error {
	kafka := dag.Kafka()

	_, err := kafka.Container().
		WithServiceBinding("kafka", kafka.Service()).
		WithExec([]string{"kafka-topics.sh", "--list", "--bootstrap-server", "kafka:9092"}).
		Sync(ctx)

	return err
}
