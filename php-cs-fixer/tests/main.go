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
	// Do something here

	return nil
}
