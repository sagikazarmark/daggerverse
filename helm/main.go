// The package manager for Kubernetes.
package main

import (
	"context"
	"fmt"
	"path"

	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "alpine/helm"

type Helm struct {
	// +private
	Ctr *Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	// +optional
	version string,

	// Custom image reference in "repository:tag" format to use as a base container.
	// +optional
	image string,

	// Custom container to use as a base container.
	// +optional
	container *Container,
) *Helm {
	var ctr *Container

	if version != "" {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	} else if image != "" {
		ctr = dag.Container().From(image)
	} else if container != nil {
		ctr = container
	} else {
		ctr = dag.Container().From(defaultImageRepository)
	}

	// Disable cache mounts for now.
	// Need to figure out if they are needed at all.
	// Note: helm registry auth is stored in ~/.config/helm/registry/config.json (within helm config dir)
	// ctr = ctr..
	// 	// TODO: run as non-root
	// 	WithUser("root").
	// 	WithMountedCache("/root/.cache/helm", dag.CacheVolume("helm-cache")).
	// 	WithMountedCache("/root/.helm", dag.CacheVolume("helm-root")).
	// 	WithMountedCache("/root/.config/helm", dag.CacheVolume("helm-config"))

	return &Helm{ctr}
}

func (m *Helm) Container() *Container {
	return m.Ctr
}

// Lint a Helm chart directory.
func (m *Helm) Lint(
	ctx context.Context,

	// A directory containing a Helm chart.
	chart *Directory,
) (*Container, error) {
	chartMetadata, err := getChartMetadata(ctx, chart)
	if err != nil {
		return nil, err
	}

	const workdir = "/src"

	chartName := chartMetadata.Name
	chartPath := path.Join(workdir, chartName)

	ctr := m.Ctr.
		WithWorkdir(workdir).
		WithMountedDirectory(chartPath, chart)

	args := []string{
		"lint",
		chartPath,
	}

	return ctr.WithExec(args), nil
}

// Build a Helm chart package.
func (m *Helm) Package(
	ctx context.Context,

	// A directory containing a Helm chart.
	chart *Directory,

	// Set the appVersion on the chart to this version.
	// +optional
	appVersion string,

	// Set the version on the chart to this semver version.
	// +optional
	version string,

	// Update dependencies from "Chart.yaml" to dir "charts/" before packaging.
	// +optional
	dependencyUpdate bool,
) (*File, error) {
	chartMetadata, err := getChartMetadata(ctx, chart)
	if err != nil {
		return nil, err
	}

	const workdir = "/src"

	chartName := chartMetadata.Name
	chartPath := path.Join(workdir, chartName)

	ctr := m.Ctr.
		WithWorkdir(workdir).
		WithMountedDirectory(chartPath, chart)

	args := []string{
		"package",

		"--destination", "/out",
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

	return ctr.File(fmt.Sprintf("/out/%s-%s.tgz", chartName, chartVersion)), nil
}

// Authenticate to an OCI registry.
func (m *Helm) Login(
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
) (*Helm, error) {
	pass, err := password.Plaintext(ctx)
	if err != nil {
		return nil, err
	}

	args := []string{
		"registry",
		"login",
		host,
		"--username", username,
		"--password", pass,
	}

	if insecure {
		args = append(args, "--insecure")
	}

	return &Helm{m.Ctr.WithExec(args)}, nil
}

// Remove credentials stored for an OCI registry.
func (m *Helm) Logout(host string) *Helm {
	args := []string{
		"registry",
		"logout",
		host,
	}

	return &Helm{m.Ctr.WithExec(args)}
}

// Push a Helm chart package to an OCI registry.
func (m *Helm) Push(
	// Packaged Helm chart.
	pkg *File,

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
) *Container {
	const workdir = "/src"

	chartPath := path.Join(workdir, "chart.tgz")

	ctr := m.Ctr.
		WithWorkdir(workdir).
		WithMountedFile(chartPath, pkg)

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

	return ctr.WithExec(args)
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
