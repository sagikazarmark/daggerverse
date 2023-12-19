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

type Helm struct{}

// Specify which version (image tag) of Helm to use from the official image repository on Docker Hub.
func (m *Helm) FromVersion(version string) *Base {
	return &Base{wrapContainer(dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version)))}
}

// Specify a custom image reference in "repository:tag" format.
func (m *Helm) FromImage(ref string) *Base {
	return &Base{wrapContainer(dag.Container().From(ref))}
}

// Specify a custom container.
func (m *Helm) FromContainer(ctr *Container) *Base {
	return &Base{wrapContainer(ctr)}
}

func defaultContainer() *Base {
	return &Base{wrapContainer(dag.Container().From(defaultImageRepository))}
}

func wrapContainer(c *Container) *Container {
	return c

	// Disable cache mounts for now.
	// Need to figure out if they are needed at all.
	// Note: helm registry auth is stored in ~/.config/helm/registry/config.json (within helm config dir)
	// return c.
	// 	// TODO: run as non-root
	// 	WithUser("root").
	// 	WithMountedCache("/root/.cache/helm", dag.CacheVolume("helm-cache")).
	// 	WithMountedCache("/root/.helm", dag.CacheVolume("helm-root")).
	// 	WithMountedCache("/root/.config/helm", dag.CacheVolume("helm-config"))
}

// Return the default container.
func (m *Helm) Container() *Container {
	return defaultContainer().Container()
}

type Base struct {
	Ctr *Container
}

// Return the underlying container.
func (m *Base) Container() *Container {
	return m.Ctr
}

// Lint a Helm chart directory.
func (m *Base) Lint(ctx context.Context, chart *Directory, appVersion Optional[string], version Optional[string]) (*Container, error) {
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
func (m *Base) Package(ctx context.Context, chart *Directory, appVersion Optional[string], version Optional[string]) (*File, error) {
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

	args = append(args, chartPath)

	ctr = ctr.WithExec(args)

	return ctr.File(fmt.Sprintf("/out/%s-%s.tgz", chartName, chartVersion)), nil
}

// Authenticate to an OCI registry.
func (m *Base) Login(ctx context.Context, host string, username string, password *Secret, insecure Optional[bool]) (*Base, error) {
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

	return &Base{m.Ctr.WithExec(args)}, nil
}

// Remove credentials stored for an OCI registry.
func (m *Base) Logout(host string) *Base {
	args := []string{
		"registry",
		"logout",
		host,
	}

	return &Base{m.Ctr.WithExec(args)}
}

// Push a Helm chart package to an OCI registry.
func (m *Base) Push(pkg *File, registry string, plainHttp Optional[bool], insecureSkipTlsVerify Optional[bool], caFile Optional[*File], certFile Optional[*File], keyFile Optional[*File]) *Container {
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
