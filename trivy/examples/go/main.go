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

	p.Go(m.Trivy_ScanContainer)

	return p.Wait()
}

// This example showcases how to initialize the Trivy module.
func (m *Examples) Trivy_New() {
	dag.Trivy(dagger.TrivyOpts{
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

// This example showcases how to scan a container with Trivy.
func (m *Examples) Trivy_ScanContainer(ctx context.Context) error {
	// Initialize Trivy module
	trivy := m.trivy()

	// Grab or build a container
	container := dag.Container().From("alpine:latest")

	// Scan the container
	report := trivy.Container(container)

	// Grab the the report output
	{
		output, err := report.Output(ctx, dagger.TrivyScanOutputOpts{
			// This is the default, but you can pass a format to this function as well
			Format: "table",
		})
		if err != nil {
			return err
		}

		_ = output
	}

	// Grab the report as a file
	{
		output, err := report.Report("json").Sync(ctx)
		if err != nil {
			return err
		}

		_ = output
	}

	return nil
}
