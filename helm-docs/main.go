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
	version Optional[string],

	// Custom image reference in "repository:tag" format to use as a base container.
	image Optional[string],

	// Custom container to use as a base container.
	container Optional[*Container],
) *HelmDocs {
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

	return &HelmDocs{
		Ctr: ctr,
	}
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
	templates Optional[[]*File],

	// Order in which to sort the values table ("alphanum" or "file"). (default "alphanum")
	sortValuesOrder Optional[string],
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

	if files, ok := templates.Get(); ok {
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
