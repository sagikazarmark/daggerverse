// Find vulnerabilities, misconfigurations, secrets, SBOM in containers, Kubernetes, code repositories, clouds and more.

package main

import (
	"context"
	"dagger/trivy/internal/dagger"
	"time"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "aquasec/trivy"

type Trivy struct {
	// +private
	Ctr *dagger.Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	version string,

	// Custom container to use as a base container. Takes precedence over version.
	//
	// +optional
	container *dagger.Container,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,

	// Persist Trivy cache between runs.
	//
	// +optional
	cache *dagger.CacheVolume,

	// OCI repository to retrieve trivy-db from. (default "ghcr.io/aquasecurity/trivy-db:2")
	//
	// +optional
	databaseRepository string,

	// Warm the vulnerability database cache.
	//
	// +optional
	warmDatabaseCache bool,
) *Trivy {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(defaultImageRepository + ":" + version)
	}

	container = container.
		With(withConfigFunc(config)).

		// Suppress progress bars
		WithEnvVariable("TRIVY_NO_PROGRESS", "true").

		// No hacking!
		WithoutEnvVariable("TRIVY_FORMAT").
		WithoutEnvVariable("TRIVY_OUTPUT")

	if cache != nil {
		const cachePath = "/tmp/cache/trivy"

		container = container.
			// Make sure parent container has no custom cache settings
			WithEnvVariable("TRIVY_CACHE_BACKEND", "fs").
			WithEnvVariable("TRIVY_CACHE_DIR", cachePath).
			WithMountedCache(cachePath, cache)
	}

	if databaseRepository != "" {
		container = container.WithEnvVariable("TRIVY_DB_REPOSITORY", databaseRepository)
	}

	if warmDatabaseCache {
		container = container.
			WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)). // We want to keep the database up-to-date
			WithExec([]string{"trivy", "image", "--download-db-only"})
	}

	return &Trivy{
		Ctr: container,
	}
}

func withConfigFunc(config *dagger.File) func(*dagger.Container) *dagger.Container {
	return func(c *dagger.Container) *dagger.Container {
		if config != nil {
			const configPath = "/work/trivy.yaml"

			c = c.
				WithEnvVariable("TRIVY_CONFIG", configPath).
				WithMountedFile(configPath, config)
		}

		return c
	}
}

type Scan struct {
	// +private
	Container *dagger.Container

	// +private
	Command *ScanCommand
}

// TODO: enabled report format enum once it's fixed
// type ReportFormat string
//
// const (
// 	Table      ReportFormat = "table"
// 	JSON       ReportFormat = "json"
// 	Template   ReportFormat = "template"
// 	SARIF      ReportFormat = "sarif"
// 	CycloneDX  ReportFormat = "cyclonedx"
// 	SPDX       ReportFormat = "spdx"
// 	SPDXJSON   ReportFormat = "spdx_json"
// 	GitHub     ReportFormat = "github"
// 	CosignVuln ReportFormat = "cosign_vuln"
// )

// Get the scan results.
func (m *Scan) Output(
	ctx context.Context,

	// Trivy report format.
	//
	// +optional
	format string,
) (string, error) {
	if format != "" {
		m.Command.Format = format
	}

	return m.Container.
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec(m.Command.args()).
		Stdout(ctx)
}

// Get the scan report as a file.
func (m *Scan) Report(
	ctx context.Context,

	// Trivy report format.
	format string,
) *dagger.File {
	reportPath := "/work/report"

	cmd := m.Command

	if format != "" {
		reportPath += "." + string(format)

		cmd.Format = format
	}

	cmd.Output = reportPath

	return m.Container.
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec(cmd.args()).
		File(reportPath)
}

type ScanCommand struct {
	Command string

	Format string
	Output string

	Args []string
}

func (c *ScanCommand) args() []string {
	args := []string{"trivy", c.Command}

	if c.Format != "" {
		args = append(args, "--format", c.Format)
	}

	if c.Output != "" {
		args = append(args, "--output", c.Output)
	}

	args = append(args, c.Args...)

	return args
}

// Scan a container.
func (m *Trivy) Container(
	// Image container to scan.
	container *dagger.Container,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,
) *Scan {
	imagePath := "/work/image.tar"

	cmd := &ScanCommand{
		Command: "image",
		Args: []string{
			"--input", imagePath,
		},
	}

	ctr := m.Ctr.
		With(withConfigFunc(config)).
		WithMountedFile(imagePath, container.AsTarball())

	return &Scan{
		Container: ctr,
		Command:   cmd,
	}
}
