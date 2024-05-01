package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Build)

	return p.Wait()
}

func (m *Tests) Build(ctx context.Context) error {
	binary := dag.Xk6().Build()

	_, err := dag.Container().
		From("alpine").
		WithMountedFile("/usr/local/bin/k6", binary).
		WithExec([]string{"k6", "version"}).
		Sync(ctx)

	return err
}
