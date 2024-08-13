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

	p.Go(m.Trivy_Output)
	p.Go(m.Trivy_Image)
	p.Go(m.Trivy_ImageFile)
	p.Go(m.Trivy_Container)
	p.Go(m.Trivy_Helm)

	return p.Wait()
}

// This example showcases how to initialize the Trivy module.
func (m *Examples) Trivy_New() {
	dag.Trivy(dagger.TrivyOpts{
		// Persist cache between runs
		Cache: dag.CacheVolume("trivy"),

		// Preheat vulnerability database cache
		WarmDatabaseCache: true,
	})
}

func (m *Examples) trivy() *dagger.Trivy {
	return dag.Trivy(dagger.TrivyOpts{
		// Persist cache between runs
		Cache: dag.CacheVolume("trivy"),

		// Preheat vulnerability database cache
		WarmDatabaseCache: true,
	})
}

// This example showcases how to initialize the Trivy module.
func (m *Examples) Trivy_Output(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.trivy()

	// Scan resources
	scans := []*dagger.TrivyScan{
		trivy.Container(dag.Container().From("alpine:latest")),
		trivy.HelmChart(dag.Helm().Create("foo").Package().File()),
	}

	// Grab the the report output
	{
		output, err := scans[0].Output(ctx, dagger.TrivyScanOutputOpts{
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
		output, err := scans[1].Report("json").Sync(ctx)
		if err != nil {
			return err
		}

		_ = output
	}

	return nil
}

// This example showcases how to scan an image (pulled from a remote repository) with Trivy.
func (m *Examples) Trivy_Image(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.trivy()

	// Scan the image
	scan := trivy.Image("alpine:latest")

	// See "Output" example.
	_, err := scan.Output(ctx)
	if err != nil {
		return err
	}

	return nil
}

// This example showcases how to scan an image archive with Trivy.
func (m *Examples) Trivy_ImageFile(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.trivy()

	// Scan the image file (using a container here for simplicity, but any image file will do)
	scan := trivy.ImageFile(dag.Container().From("alpine:latest").AsTarball())

	// See "Output" example.
	_, err := scan.Output(ctx)
	if err != nil {
		return err
	}

	return nil
}

// This example showcases how to scan a container with Trivy.
func (m *Examples) Trivy_Container(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.trivy()

	// Grab or build a container
	container := dag.Container().From("alpine:latest")

	// Scan the container
	scan := trivy.Container(container)

	// See "Output" example.
	_, err := scan.Output(ctx)
	if err != nil {
		return err
	}

	return nil
}

// This example showcases how to scan a Helm chart with Trivy.
func (m *Examples) Trivy_Helm(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.trivy()

	// Grab or build a Helm chart package
	chart := dag.Helm().Create("foo").Package()

	// Scan the Helm chart
	scan := trivy.HelmChart(chart.File())

	// See "Output" example.
	_, err := scan.Output(ctx)
	if err != nil {
		return err
	}

	return nil
}
