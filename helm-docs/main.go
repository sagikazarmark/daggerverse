// A tool for automatically generating markdown documentation for helm charts.
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

type HelmDocs struct {
	// +private
	Container *Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	version string,

	// Custom container to use as a base container.
	//
	// +optional
	container *Container,
) *HelmDocs {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	return &HelmDocs{
		Container: container,
	}
}

// Generate markdown documentation for Helm charts from requirements and values files.
func (m *HelmDocs) Generate(
	ctx context.Context,

	// A directory containing a Helm chart.
	chart *Directory,

	// A list of Go template files to use for rendering the documentation.
	//
	// +optional
	templates []*File,

	// Order in which to sort the values table ("alphanum" or "file"). (default "alphanum")
	//
	// +optional
	sortValuesOrder string,
) (*File, error) {
	chartName, err := getChartName(ctx, chart)
	if err != nil {
		return nil, err
	}

	chartPath := path.Join("/work/charts", chartName)

	args := []string{
		// Technically this is not needed, but let's add it anyway
		"--chart-search-root", "/work/charts",

		"--chart-to-generate", chartPath,
		"--output-file", "README.out.md",

		// "--log-level", "trace",
	}

	if sortValuesOrder != "" {
		args = append(args, "--sort-values-order", sortValuesOrder)
	}

	return m.Container.
		WithWorkdir("/work").
		WithMountedDirectory(chartPath, chart).
		With(func(c *Container) *Container {
			// TODO: use shell and list files inside template directory?
			for i, template := range templates {
				templatePath := fmt.Sprintf("/work/templates/template-%d", i)

				args = append(args, "--template-files", templatePath)
				c = c.WithMountedFile(templatePath, template)
			}

			return c
		}).
		WithExec(args).
		File(path.Join(chartPath, "README.out.md")), nil
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
