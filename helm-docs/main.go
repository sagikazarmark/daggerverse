package main

import (
	"fmt"
	"path"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "jnorwood/helm-docs"

type HelmDocs struct{}

// Specify which version (image tag) of Kafka to use from the official image repository on Docker Hub.
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

func (m *HelmDocs) Generate(chartName string, chart *Directory, templateFiles Optional[[]*File]) *File {
	return defaultContainer().Generate(
		chartName,
		chart,
		templateFiles,
	)
}

type Base struct {
	Ctr *Container
}

func (m *Base) Generate(chartName string, chart *Directory, templateFiles Optional[[]*File]) *File {
	// TODO: sanitize chart name

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
	}

	if files, ok := templateFiles.Get(); ok {
		for i, file := range files {
			args = append(args, "--template-files", fmt.Sprintf("../../templates/template-%d", i))
			ctr = ctr.WithFile(fmt.Sprint("/src/templates/template-%d", i), file)
		}
	}

	ctr = ctr.WithExec(args)

	return ctr.File(path.Join(chartPath, "README.out.md"))
}
