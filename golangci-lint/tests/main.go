package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Run)

	return p.Wait()
}

func (m *Tests) Run(ctx context.Context) error {
	_, err := dag.GolangciLint().
		Run(dag.CurrentModule().Source().Directory("./testdata")).
		Sync(ctx)

	return err
}
