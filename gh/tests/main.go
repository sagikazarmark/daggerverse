package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Help)

	return p.Wait()
}

func (m *Tests) Help(ctx context.Context) error {
	_, err := dag.Gh().Run("--help").Sync(ctx)

	return err
}
