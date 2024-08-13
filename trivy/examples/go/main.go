// Examples for Trivy module

package main

import (
	"context"
	"dagger/trivy/examples/go/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

type Examples struct {
	// +private
	Trivy *dagger.Trivy
}

func New() *Examples {
	return &Examples{
		Trivy: dag.Trivy(dagger.TrivyOpts{
			// Persist cache between runs
			Cache: dag.CacheVolume("trivy"),

			// Preheat vulnerability database cache
			WarmDatabaseCache: true,
		}),
	}
}

func (m *Examples) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Trivy_Output)
	p.Go(m.Trivy_Config)
	p.Go(m.Trivy_Image)
	p.Go(m.Trivy_ImageFile)
	p.Go(m.Trivy_Container)
	p.Go(m.Trivy_Helm)
	p.Go(m.Trivy_Filesystem)
	p.Go(m.Trivy_Rootfs)
	p.Go(m.Trivy_Binary)

	return p.Wait()
}

func output(ctx context.Context, scan *dagger.TrivyScan) error {
	_, err := scan.Output(ctx)
	if err != nil {
		return err
	}

	return nil
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

// This example showcases how to initialize the Trivy module.
func (m *Examples) Trivy_Output(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.Trivy

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

// This example showcases how to pass configuration to the Trivy module.
func (m *Examples) Trivy_Config(ctx context.Context) error {
	// Initialize Trivy module with custom configuration...
	trivy := dag.Trivy(dagger.TrivyOpts{
		Config: dag.CurrentModule().Source().File("trivy.yaml"),
	})

	// ...or pass it directly to the scan
	scan := trivy.Image("alpine:latest", dagger.TrivyImageOpts{
		Config: dag.CurrentModule().Source().File("trivy.yaml"),
	})

	// See "Output" example.
	return output(ctx, scan)
}

// This example showcases how to scan an image (pulled from a remote repository) with Trivy.
func (m *Examples) Trivy_Image(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.Trivy

	// Scan the image
	scan := trivy.Image("alpine:latest")

	// See "Output" example.
	return output(ctx, scan)
}

// This example showcases how to scan an image archive with Trivy.
func (m *Examples) Trivy_ImageFile(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.Trivy

	// Scan the image file (using a container here for simplicity, but any image file will do)
	scan := trivy.ImageFile(dag.Container().From("alpine:latest").AsTarball())

	// See "Output" example.
	return output(ctx, scan)
}

// This example showcases how to scan a container with Trivy.
func (m *Examples) Trivy_Container(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.Trivy

	// Grab or build a container
	container := dag.Container().From("alpine:latest")

	// Scan the container
	scan := trivy.Container(container)

	// See "Output" example.
	return output(ctx, scan)
}

// This example showcases how to scan a Helm chart with Trivy.
func (m *Examples) Trivy_Helm(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.Trivy

	// Grab or build a Helm chart package
	chart := dag.Helm().Create("foo").Package()

	// Scan the Helm chart
	scan := trivy.HelmChart(chart.File())

	// See "Output" example.
	return output(ctx, scan)
}

// This example showcases how to scan a filesystem with Trivy.
func (m *Examples) Trivy_Filesystem(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.Trivy

	// Grab a directory
	directory := dag.Git("https://github.com/sagikazarmark/daggerverse.git").Head().Tree()

	// Scan the filesystem
	scan := trivy.Filesystem(directory)

	// See "Output" example.
	return output(ctx, scan)
}

// This example showcases how to scan a rootfs with Trivy.
func (m *Examples) Trivy_Rootfs(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.Trivy

	// Grab the rootfs of a container
	rootfs := dag.Container().From("alpine:latest").Rootfs()

	// Scan the rootfs
	scan := trivy.Rootfs(rootfs)

	// See "Output" example.
	return output(ctx, scan)
}

// This example showcases how to scan a binary with Trivy.
func (m *Examples) Trivy_Binary(ctx context.Context) error {
	// Initialize Trivy module
	// See "New" example.
	trivy := m.Trivy

	// Grab a binary file
	binary := dag.Container().From("alpine:latest").File("/usr/bin/env")

	// Scan the binary
	scan := trivy.Binary(binary)

	// See "Output" example.
	return output(ctx, scan)
}
