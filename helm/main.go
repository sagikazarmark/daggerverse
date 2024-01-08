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
	version Optional[string],

	// Custom image reference in "repository:tag" format to use as a base container.
	image Optional[string],

	// Custom container to use as a base container.
	container Optional[*Container],
) *Helm {
	var ctr *Container

	if v, ok := version.Get(); ok {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, v))
	} else if i, ok := image.Get(); ok {
		ctr = dag.Container().From(i)
	} else if c, ok := container.Get(); ok {
		ctr = c
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

	return &Helm{
		Ctr: ctr,
	}
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
	appVersion Optional[string],

	// Set the version on the chart to this semver version.
	version Optional[string],

	// Update dependencies from "Chart.yaml" to dir "charts/" before packaging.
	dependencyUpdate Optional[bool],
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

	if v, ok := appVersion.Get(); ok {
		args = append(args, "--app-version", v)
	}

	chartVersion := chartMetadata.Version
	if v, ok := version.Get(); ok {
		args = append(args, "--version", v)
		chartVersion = v
	}

	if v, ok := dependencyUpdate.Get(); ok && v {
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
	insecure Optional[bool],
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

	if v, ok := insecure.Get(); ok && v {
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
	plainHttp Optional[bool],

	// Skip tls certificate checks for the chart upload.
	insecureSkipTlsVerify Optional[bool],

	// Verify certificates of HTTPS-enabled servers using this CA bundle.
	caFile Optional[*File],

	// Identify registry client using this SSL certificate file.
	certFile Optional[*File],

	// Identify registry client using this SSL key file.
	keyFile Optional[*File],
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

	if v, ok := plainHttp.Get(); ok && v {
		args = append(args, "--plain-http")
	}

	if v, ok := insecureSkipTlsVerify.Get(); ok && v {
		args = append(args, "--insecure-skip-tls-verify")
	}

	if v, ok := caFile.Get(); ok {
		ctr = ctr.WithMountedFile("/etc/helm/ca.pem", v)
		args = append(args, "--ca-file", "/etc/helm/ca.pem")
	}

	if v, ok := certFile.Get(); ok {
		ctr = ctr.WithMountedFile("/etc/helm/cert.pem", v)
		args = append(args, "--cert-file", "/etc/helm/cert.pem")
	}

	if v, ok := keyFile.Get(); ok {
		ctr = ctr.WithMountedFile("/etc/helm/key.pem", v)
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
