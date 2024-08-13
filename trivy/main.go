// Find vulnerabilities, misconfigurations, secrets, SBOM in containers, Kubernetes, code repositories, clouds and more.

package main

import (
	"context"
	"dagger/trivy/internal/dagger"
	"errors"
	"strings"
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

// Supported Trivy scan kinds.
type ScanKind string

const (
	Image      ScanKind = "image"
	Filesystem ScanKind = "filesystem"
	Rootfs     ScanKind = "rootfs"
	Config     ScanKind = "config"
	SBOM       ScanKind = "sbom"
)

type Scan struct {
	// +private
	Container *dagger.Container

	// +private
	Config *dagger.File

	// +private
	Kind ScanKind

	// +private
	Source *dagger.Directory

	// +private
	Target string

	// +private
	Args []string
}

func (m *Scan) container(extraArgs []string) *dagger.Container {
	container := m.Container.
		WithWorkdir("/work") // default workdir

	args := []string{"trivy", string(m.Kind)}

	if m.Config != nil {
		container = container.WithMountedFile("/work/trivy.yaml", m.Config)
		args = append(args, "--config", "/work/trivy.yaml")
	}

	if m.Source != nil {
		container = container.
			WithMountedDirectory("/work/source", m.Source).
			WithWorkdir("/work/source")
	}

	args = append(args, m.Args...)
	args = append(args, extraArgs...)

	if m.Target != "" {
		args = append(args, m.Target)
	}

	return container.
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec(args)
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
	var args []string

	if format != "" {
		args = append(args, "--format", format)
	}

	return m.container(args).Stdout(ctx)
}

// Get the scan report as a file.
func (m *Scan) Report(
	ctx context.Context,

	// Trivy report format.
	format string,
) *dagger.File {
	reportPath := "/work/report"

	var args []string

	if format != "" {
		args = append(args, "--format", format)
		reportPath += "." + string(format)
	}

	args = append(args, "--output", reportPath)

	return m.container(args).File(reportPath)
}

// Scan a container image.
//
// See https://aquasecurity.github.io/trivy/latest/docs/target/container_image/ for more information.
func (m *Trivy) Image(
	// Name of the image to scan.
	image string,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,
) *Scan {
	return &Scan{
		Container: m.Ctr,
		Config:    config,
		Kind:      Image,
		Target:    image,
	}
}

// Scan a container image file.
//
// See https://aquasecurity.github.io/trivy/latest/docs/target/container_image/ for more information.
func (m *Trivy) ImageFile(
	// Input file to the image (to use instead of pulling).
	image *dagger.File,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,
) *Scan {
	const inputName = "image.tar"

	source := dag.Directory().WithFile(inputName, image)

	return &Scan{
		Container: m.Ctr,
		Config:    config,
		Kind:      Image,
		Source:    source,
		Args:      []string{"--input", inputName},
	}
}

// Scan a container.
//
// See https://aquasecurity.github.io/trivy/latest/docs/target/container_image/ for more information.
func (m *Trivy) Container(
	// Image container to scan.
	container *dagger.Container,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,
) *Scan {
	return m.ImageFile(container.AsTarball(), config)
}

// Scan a Helm chart.
func (m *Trivy) HelmChart(
	ctx context.Context,

	// Helm chart package to scan.
	chart *dagger.File,

	// Inline values for the Helm chart (equivalent of --set parameter of the helm install command).
	//
	// +optional
	set []string,

	// Inline values for the Helm chart (equivalent of --set-string parameter of the helm install command).
	//
	// +optional
	setString []string,

	// Values files for the Helm chart (equivalent of --values parameter of the helm install command).
	//
	// +optional
	values []*dagger.File,

	// Kubernetes version used for Capabilities.KubeVersion.
	//
	// +optional
	kubeVersion string,

	// Available API versions used for Capabilities.APIVersions.
	//
	// +optional
	apiVersions []string,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,
) (*Scan, error) {
	const input = "chart.tgz"

	source := dag.Directory().WithFile(input, chart)

	var args []string

	if len(set) > 0 {
		args = append(args, "--helm-set", strings.Join(set, ","))
	}

	if len(setString) > 0 {
		args = append(args, "--helm-set-string", strings.Join(setString, ","))
	}

	if len(values) > 0 {
		dir := dag.Directory().WithFiles("", values)

		entries, err := dir.Entries(ctx)
		if err != nil {
			return nil, err
		}

		for i, v := range entries {
			entries[i] = "values/" + v
		}

		args = append(args, "--helm-values", strings.Join(entries, ","))
		source = source.WithDirectory("values", dir)
	}

	if kubeVersion != "" {
		args = append(args, "--helm-kube-version", kubeVersion)
	}

	if len(apiVersions) > 0 {
		args = append(args, "--helm-api-versions", strings.Join(apiVersions, ","))
	}

	return &Scan{
		Container: m.Ctr,
		Config:    config,
		Kind:      Config,
		Source:    source,
		Target:    input,
	}, nil
}

// Scan a filesystem.
//
// See https://aquasecurity.github.io/trivy/latest/docs/target/filesystem/ for more information.
func (m *Trivy) Filesystem(
	// Directory to scan.
	directory *dagger.Directory,

	// Subpath within the directory to scan.
	//
	// +optional
	// +default="."
	target string,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,
) *Scan {
	return &Scan{
		Container: m.Ctr,
		Config:    config,
		Kind:      Filesystem,
		Source:    directory,
		Target:    target,
	}
}

// Scan a root filesystem.
//
// See https://aquasecurity.github.io/trivy/latest/docs/target/rootfs/ for more information.
func (m *Trivy) Rootfs(
	// Directory to scan.
	directory *dagger.Directory,

	// Subpath within the directory to scan.
	//
	// +optional
	// +default="."
	target string,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,
) *Scan {
	return &Scan{
		Container: m.Ctr,
		Config:    config,
		Kind:      Rootfs,
		Source:    directory,
		Target:    target,
	}
}

// Scan a binary.
//
// This is a convenience method to scan a binary file that normally falls under the rootfs target.
//
// See https://aquasecurity.github.io/trivy/latest/docs/target/rootfs/ for more information.
func (m *Trivy) Binary(
	ctx context.Context,

	// Binary to scan.
	file *dagger.File,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,
) (*Scan, error) {
	name, err := file.Name(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: is this even possible?
	if name == "" {
		name = "binary"
	}

	dir := dag.Directory().WithFile(name, file)

	return m.Rootfs(dir, name, config), nil
}

// Scan an SBOM.
//
// See https://aquasecurity.github.io/trivy/latest/docs/target/sbom/ for more information.
func (m *Trivy) Sbom(
	ctx context.Context,

	// SBOM to scan.
	sbom *dagger.File,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,
) (*Scan, error) {
	name, err := sbom.Name(ctx)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, errors.New("sbom file has no name")
	}

	source := dag.Directory().WithFile(name, sbom)

	return &Scan{
		Container: m.Ctr,
		Config:    config,
		Kind:      SBOM,
		Source:    source,
		Target:    name,
	}, nil
}
