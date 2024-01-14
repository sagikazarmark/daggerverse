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
) *HelmDocs {
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

	return &HelmDocs{ctr}
}

func (m *HelmDocs) Container() *Container {
	return m.Ctr
}

// Generate markdown documentation for Helm charts from requirements and values files.
func (m *HelmDocs) Generate(
	ctx context.Context,

	// A directory containing a Helm chart.
	chart *Directory,

	// A list of Go template files to use for rendering the documentation.
	// +optional
	templates []*File,

	// Order in which to sort the values table ("alphanum" or "file"). (default "alphanum")
	// +optional
	sortValuesOrder string,
) (*File, error) {
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
		"--chart-search-root", "/src/charts",

		"--chart-to-generate", chartPath,
		"--output-file", "README.out.md",

		// "--log-level", "trace",
	}

	for i, template := range templates {
		templatePath := fmt.Sprint("/src/templates/template-%d", i)

		args = append(args, "--template-files", templatePath)
		ctr = ctr.WithMountedFile(templatePath, template)
	}

	if sortValuesOrder != "" {
		args = append(args, "--sort-values-order", sortValuesOrder)
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
