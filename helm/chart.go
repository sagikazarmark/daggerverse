package main

import (
	"context"
	"fmt"
	"path"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"
)

// Returns a Helm chart from a source directory.
func (m *Helm) Chart(
	// A directory containing a Helm chart.
	source *Directory,
) *Chart {
	return &Chart{
		Directory: source,
		Container: m.Ctr,
	}
}

// A Helm chart.
type Chart struct {
	Directory *Directory

	// +private
	Container *Container
}

// Lint a Helm chart.
func (c *Chart) Lint(ctx context.Context) (*Container, error) {
	chartMetadata, err := getChartMetadata(ctx, c.Directory)
	if err != nil {
		return nil, err
	}

	const workdir = "/work"

	chartName := chartMetadata.Name
	chartPath := path.Join(workdir, chartName)

	ctr := c.Container.
		WithWorkdir(workdir).
		WithMountedDirectory(chartPath, c.Directory)

	args := []string{
		"lint",
		chartPath,
	}

	return ctr.WithExec(args), nil
}

// A Helm chart package.
type Package struct {
	File *File

	// +private
	Container *Container
}

// Build a Helm chart package.
func (c *Chart) Package(
	ctx context.Context,

	// Set the appVersion on the chart to this version.
	// +optional
	appVersion string,

	// Set the version on the chart to this semver version.
	// +optional
	version string,

	// Update dependencies from "Chart.yaml" to dir "charts/" before packaging.
	// +optional
	dependencyUpdate bool,
) (*Package, error) {
	chartMetadata, err := getChartMetadata(ctx, c.Directory)
	if err != nil {
		return nil, err
	}

	const workdir = "/work"

	chartName := chartMetadata.Name
	chartPath := filepath.Join(workdir, chartName)

	ctr := c.Container.
		WithWorkdir(workdir).
		WithMountedDirectory(chartPath, c.Directory)

	args := []string{
		"package",

		"--destination", workdir,
	}

	if appVersion != "" {
		args = append(args, "--app-version", appVersion)
	}

	chartVersion := chartMetadata.Version
	if version != "" {
		args = append(args, "--version", version)
		chartVersion = version
	}

	if dependencyUpdate {
		args = append(args, "--dependency-update")
	}

	args = append(args, chartPath)

	ctr = ctr.WithExec(args)

	return &Package{
		File:      ctr.File(filepath.Join(workdir, fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))),
		Container: c.Container,
	}, nil
}

// Authenticate to an OCI registry.
func (p *Package) WithRegistryAuth(
	ctx context.Context,

	// Host of the OCI registry.
	host string,

	// Registry username.
	username string,

	// Registry password.
	password *Secret,

	// Allow connections to TLS registry without certs.
	// +optional
	insecure bool,
) (*Package, error) {
	ctr, err := login(ctx, p.Container, host, username, password, insecure)
	if err != nil {
		return nil, err
	}

	return &Package{
		File:      p.File,
		Container: ctr,
	}, nil
}

// Remove credentials stored for an OCI registry.
func (p *Package) WithoutRegistryAuth(host string) *Package {
	return &Package{
		File:      p.File,
		Container: logout(p.Container, host),
	}
}

// Publishes this Helm chart package to an OCI registry.
func (p *Package) Publish(
	ctx context.Context,

	// OCI registry to push to (including the path except the chart name).
	registry string,

	// Use insecure HTTP connections for the chart upload.
	// +optional
	plainHttp bool,

	// Skip tls certificate checks for the chart upload.
	// +optional
	insecureSkipTlsVerify bool,

	// Verify certificates of HTTPS-enabled servers using this CA bundle.
	// +optional
	caFile *File,

	// Identify registry client using this SSL certificate file.
	// +optional
	certFile *File,

	// Identify registry client using this SSL key file.
	// +optional
	keyFile *File,
) error {
	const workdir = "/work"

	chartPath := path.Join(workdir, "chart.tgz")

	ctr := p.Container.
		WithWorkdir(workdir).
		WithMountedFile(chartPath, p.File)

	args := []string{
		"push",

		chartPath,
		registry,
	}

	if plainHttp {
		args = append(args, "--plain-http")
	}

	if insecureSkipTlsVerify {
		args = append(args, "--insecure-skip-tls-verify")
	}

	if caFile != nil {
		ctr = ctr.WithMountedFile("/etc/helm/ca.pem", caFile)
		args = append(args, "--ca-file", "/etc/helm/ca.pem")
	}

	if certFile != nil {
		ctr = ctr.WithMountedFile("/etc/helm/cert.pem", certFile)
		args = append(args, "--cert-file", "/etc/helm/cert.pem")
	}

	if keyFile != nil {
		ctr = ctr.WithMountedFile("/etc/helm/key.pem", keyFile)
		args = append(args, "--key-file", "/etc/helm/key.pem")
	}

	_, err := ctr.WithExec(args).Sync(ctx)

	return err
}

func getChartMetadata(ctx context.Context, c *Directory) (*chart.Metadata, error) {
	chartYaml, err := c.File("Chart.yaml").Contents(ctx)
	if err != nil {
		return nil, err
	}

	meta := new(chart.Metadata)
	err = yaml.Unmarshal([]byte(chartYaml), meta)
	if err != nil {
		return nil, err
	}

	return meta, nil
}
