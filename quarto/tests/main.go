package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Render)

	return p.Wait()
}

func (m *Tests) Render(ctx context.Context) error {
	dir := dag.CurrentModule().Source().Directory("./testdata")

	_, err := dag.Quarto().Render(dir).Directory().Sync(ctx)
	if err != nil {
		return err
	}

	return nil
}
