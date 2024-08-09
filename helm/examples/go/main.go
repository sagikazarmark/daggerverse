// Examples for Helm module

package main

import (
	"context"
	"dagger/examples/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

type Examples struct{}

// All executes all examples.
func (m *Examples) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Helm_Create)

	return p.Wait()
}

func (m *Examples) Helm_Create(ctx context.Context) error {
	chart := dag.Helm().Create("foo")

	_, err := chart.Package(dagger.HelmChartPackageOpts{
		AppVersion: "1.0.0",
		Version:    "1.0.0",
	}).File().Sync(ctx)

	return err
}
