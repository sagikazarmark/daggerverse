// Examples for Trivy module

package main

import (
	"context"
	"dagger/trivy/examples/go/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

type Examples struct{}

func (m *Examples) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Trivy_Container)

	return p.Wait()
}

func (m *Examples) Trivy_Init() {
	dag.Trivy(dagger.TrivyOpts{
		// Persist cache between runs
		Cache: dag.CacheVolume("trivy"),
	}).
		// Preheat vulnerability database cache
		DownloadDb()
}

func (m *Examples) trivy() *dagger.Trivy {
	return dag.Trivy(dagger.TrivyOpts{
		// Persist cache between runs
		Cache: dag.CacheVolume("trivy"),
	}).
		// Preheat vulnerability database cache
		DownloadDb()
}

func (m *Examples) Trivy_Container(ctx context.Context) error {
	// Initialize Trivy module
	trivy := m.trivy()

	// Grab or build a container
	container := dag.Container().From("alpine:latest")

	// Scan the container
	report := trivy.Container(container)

	// Grab the the report output
	{
		output, err := report.Output(ctx)
		if err != nil {
			return err
		}

		_ = output
	}

	// Grab the report as a file
	{
		output, err := report.Report(dagger.TrivyScanReportOpts{
			Format: "json",
		}).Sync(ctx)
		if err != nil {
			return err
		}

		_ = output
	}

	return nil
}
