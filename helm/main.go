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
	return c.
		// TODO: run as non-root
		WithUser("root").
		WithMountedCache("/root/.cache/helm", dag.CacheVolume("helm-cache")).
		WithMountedCache("/root/.helm", dag.CacheVolume("helm-root")).
		WithMountedCache("/root/.config/helm", dag.CacheVolume("helm-config"))
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
