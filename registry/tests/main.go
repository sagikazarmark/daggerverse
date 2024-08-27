package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Test)

	return p.Wait()
}

func (m *Tests) Test(ctx context.Context) error {
	registry := dag.Registry()

	_, err := dag.Container().
		From("alpine:latest").
		WithServiceBinding("registry", registry.Service()).
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{"curl", "http://registry:5000/v2"}).
		Sync(ctx)

	return err
}
