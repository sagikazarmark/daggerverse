// Find vulnerabilities, misconfigurations, secrets, SBOM in containers, Kubernetes, code repositories, clouds and more.

package main

import (
	"context"
	"dagger/trivy/internal/dagger"
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
	cmd := &ScanCommand{
		Command: "image",
		Args:    []string{image},
	}

	ctr := m.Ctr.
		With(withConfigFunc(config)).
		WithWorkdir("/work")

	return &Scan{
		Container: ctr,
		Command:   cmd,
	}
}

// Scan a container image file.
//
// See https://aquasecurity.github.io/trivy/latest/docs/target/container_image/ for more information.
func (m *Trivy) ImageFile(
	// Input file to the image (to use instead of pulling).
	input *dagger.File,

	// Trivy configuration file.
	//
	// +optional
	config *dagger.File,
) *Scan {
	const inputName = "image.tar"

	cmd := &ScanCommand{
		Command: "image",
		Args:    []string{"--input", inputName},
	}

	const workDir = "/work/source"

	ctr := m.Ctr.
		With(withConfigFunc(config)).
		WithMountedFile(workDir+"/"+inputName, input).
		WithWorkdir(workDir)

	return &Scan{
		Container: ctr,
		Command:   cmd,
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
	const workDir = "/work/source"
	const input = "chart.tgz"

	cmd := &ScanCommand{
		Command: "config",
		Args:    []string{input},
	}

	ctr := m.Ctr

	if len(set) > 0 {
		cmd.Args = append(cmd.Args, "--helm-set", strings.Join(set, ","))
	}

	if len(setString) > 0 {
		cmd.Args = append(cmd.Args, "--helm-set-string", strings.Join(setString, ","))
	}

	if len(values) > 0 {
		dir := dag.Directory().WithFiles("", values)

		entries, err := dir.Entries(ctx)
		if err != nil {
			return nil, err
		}

		for i, v := range entries {
			entries[i] = "/work/values/" + v
		}

		cmd.Args = append(cmd.Args, "--helm-values", strings.Join(entries, ","))
		ctr = ctr.WithMountedDirectory("/work/values", dir)
	}

	if kubeVersion != "" {
		cmd.Args = append(cmd.Args, "--helm-kube-version", kubeVersion)
	}

	if len(apiVersions) > 0 {
		cmd.Args = append(cmd.Args, "--helm-api-versions", strings.Join(apiVersions, ","))
	}

	ctr = ctr.
		With(withConfigFunc(config)).
		WithMountedFile(workDir+"/"+input, chart).
		WithWorkdir(workDir)

	return &Scan{
		Container: ctr,
		Command:   cmd,
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
	const workDir = "/work/source"

	cmd := &ScanCommand{
		Command: "filesystem",
		Args:    []string{target},
	}

	ctr := m.Ctr.
		With(withConfigFunc(config)).
		WithMountedDirectory(workDir, directory).
		WithWorkdir(workDir)

	return &Scan{
		Container: ctr,
		Command:   cmd,
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
	const workDir = "/work/source"

	cmd := &ScanCommand{
		Command: "rootfs",
		Args:    []string{target},
	}

	ctr := m.Ctr.
		With(withConfigFunc(config)).
		WithMountedDirectory(workDir, directory).
		WithWorkdir(workDir)

	return &Scan{
		Container: ctr,
		Command:   cmd,
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
