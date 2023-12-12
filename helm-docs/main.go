package main

import (
	"context"
	"fmt"
	"path"

	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "jnorwood/helm-docs"

type HelmDocs struct{}

// Specify which version (image tag) of helm-docs to use from the official image repository on Docker Hub.
func (m *HelmDocs) FromVersion(version string) *Base {
	return &Base{dag.Container().From(fmt.Sprintf("%s:v%s", defaultImageRepository, version))}
}

// Specify a custom image reference in "repository:tag" format.
func (m *HelmDocs) FromImage(ref string) *Base {
	return &Base{dag.Container().From(ref)}
}

// Specify a custom container.
func (m *HelmDocs) FromContainer(ctr *Container) *Base {
	return &Base{ctr}
}

func defaultContainer() *Base {
	return &Base{dag.Container().From(defaultImageRepository)}
}

func (m *HelmDocs) Generate(ctx context.Context, chart *Directory, templateFiles Optional[[]*File], sortValuesOrder Optional[string]) (*File, error) {
	return defaultContainer().Generate(
		ctx,
		chart,
		templateFiles,
		sortValuesOrder,
	)
}

type Base struct {
	Ctr *Container
}

func (m *Base) Generate(ctx context.Context, chart *Directory, templateFiles Optional[[]*File], sortValuesOrder Optional[string]) (*File, error) {
	chartName, err := getChartName(ctx, chart)
	if err != nil {
		return nil, err
	}

	chartPath := path.Join("/src/charts", chartName)

	ctr := m.Ctr.
		WithWorkdir("/src").
		WithMountedDirectory(chartPath, chart)

	args := []string{
		// Technically this is not needed, but let's add it anyway
		"--chart-search-root",
		"/src/charts",

		"--chart-to-generate",
		chartPath,
		"--output-file",
		"README.out.md",

		// "--log-level",
		// "trace",
	}

	if files, ok := templateFiles.Get(); ok {
		for i, file := range files {
			templatePath := fmt.Sprint("/src/templates/template-%d", i)

			args = append(args, "--template-files", templatePath)
			ctr = ctr.WithMountedFile(templatePath, file)
		}
	}

	if v, ok := sortValuesOrder.Get(); ok {
		args = append(args, "--sort-values-order", v)
	}

	ctr = ctr.WithExec(args)

	return ctr.File(path.Join(chartPath, "README.out.md")), nil
}

func getChartName(ctx context.Context, c *Directory) (string, error) {
	chartYaml, err := c.File("Chart.yaml").Contents(ctx)
	if err != nil {
		return "", err
	}

	y := new(chart.Metadata)
	err = yaml.Unmarshal([]byte(chartYaml), y)
	if err != nil {
		return "", err
	}

	return y.Name, nil
}
