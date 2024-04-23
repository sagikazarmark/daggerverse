// Open-source scientific and technical publishing system built on Pandoc.
package main

import (
	"context"
	"fmt"

	"golang.org/x/exp/slices"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "ghcr.io/quarto-dev/quarto"

type Quarto struct {
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
) *Quarto {
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

	return &Quarto{ctr}
}

func (m *Quarto) Container() *Container {
	return m.Ctr
}

// Render files or projects to various document types.
func (m *Quarto) Render(
	ctx context.Context,

	// Quarto source directory.
	source *Directory,

	// Input to render within the project.
	// +optional
	input string,

	// Override site-url for website or book output.
	// +optional
	siteUrl string,
) *Renderer {
	args := []string{
		"quarto", "render",
	}

	if siteUrl != "" {
		args = append(args, "--site-url", siteUrl)
	}

	if input != "" {
		args = append(args, input)
	}

	return &Renderer{
		Ctr:  m.Ctr.WithWorkdir("/work/source").WithDirectory("/work/source", source),
		Args: args,
	}
}

type Renderer struct {
	// +private
	Ctr *Container

	// +private
	Args []string
}

func (m *Renderer) run(args ...string) *Container {
	args = append(slices.Clone(m.Args), args...)

	return m.Ctr.WithExec(args)
}

// Get the output directory after rendering.
func (m *Renderer) Directory() *Directory {
	return m.run("--output-dir", "../output").Directory("/work/output")
}

// Get the output file after rendering.
func (m *Renderer) File(name string) *Directory {
	return m.run("--output", "/work/source/_site/"+name).Directory("/work/source/_site/" + name)
}
