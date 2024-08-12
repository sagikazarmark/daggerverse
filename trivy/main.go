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
			// Make sure parent container has no custom cache setting
			WithEnvVariable("TRIVY_CACHE_BACKEND", "fs").
			WithEnvVariable("TRIVY_CACHE_DIR", cachePath).
			WithMountedCache(cachePath, cache)
	}

	return &Trivy{
		Ctr: container,
	}
}

// Download vulnerability database.
//
// This is done automatically when scanning, but can be called manually to warm the cache.
// This is useful when a cache volume is used.
func (m *Trivy) DownloadDB(
	// Override the default Trivy database URL.
	//
	// +optional
	repository string,
) *Trivy {
	m.Ctr = m.Ctr.
		With(func(c *dagger.Container) *dagger.Container {
			if repository != "" {
				c = c.WithEnvVariable("TRIVY_DB_REPOSITORY", repository)
			}

			return c
		}).
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)). // We want to keep the database up-to-date
		WithExec([]string{"trivy", "image", "--download-db-only"})

	return m
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

type Report struct {
	// +private
	Container *dagger.Container

	// +private
	Args []string
}

// Return the report output.
func (m *Report) Output(ctx context.Context) (string, error) {
	return m.Container.
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec(m.Args).
		Stdout(ctx)
}

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

func (m *Report) File(
	ctx context.Context,

	// Trivy report format.
	//
	// +optional
	format string,
) *dagger.File {
	reportPath := "/work/report"

	args := m.Args

	if format != "" {
		reportPath += "." + string(format)

		args = append(args, "--format", string(format))
	}

	args = append(args, "--output", reportPath)

	return m.Container.
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec(args).
		File(reportPath)
}

// Scan a container.
func (m *Trivy) Container(
	// Image container to scan.
	container *dagger.Container,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,
) *Report {
	imagePath := "/work/image.tar"

	args := []string{"trivy", "image", "--input", imagePath}

	ctr := m.Ctr.
		With(withConfigFunc(config)).
		WithMountedFile(imagePath, container.AsTarball())

	return &Report{
		Container: ctr,
		Args:      args,
	}
}