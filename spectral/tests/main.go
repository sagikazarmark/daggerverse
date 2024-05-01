package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Lint)

	return p.Wait()
}

func (m *Tests) Lint(ctx context.Context) error {
	source := dag.CurrentModule().Source().Directory("./testdata")

	_, err := dag.Spectral().Lint([]*File{source.File("openapi.json")}, source.File(".spectral.yaml")).Sync(ctx)

	return err
}
