// Your CLI home video recorder.

package main

import (
	"dagger/vhs/internal/dagger"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "ghcr.io/charmbracelet/vhs"

type Vhs struct {
	Container *dagger.Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	version string,

	// Custom container to use as a base container.
	//
	// +optional
	container *dagger.Container,
) *Vhs {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	return &Vhs{
		Container: container,
	}
}

// Create a new tape file with example tape file contents and documentation.
func (m *Vhs) NewTape(
	// Name of the tape file to create.
	//
	// +optional
	// +default="cassette.tape"
	name string,
) *dagger.File {
	if name == "" {
		name = "cassette.tape"
	}

	return m.Container.
		WithExec([]string{"vhs", "new", name}).
		File(name)
}

// Runs a given tape file and generates its outputs.
//
// If you have source commands in your tape file, use withSource instead.
func (m *Vhs) Render(
	// Tape file to render.
	tape *dagger.File,

	// File name(s) of video output.
	//
	// +optional
	// output []string,

	// Publish your GIF to vhs.charm.sh and get a shareable URL.
	//
	// +optional
	// +default=false
	publish bool,
) *dagger.Directory {
	args := []string{"vhs", "cassette.tape"}

	// Does not seem to work at the moment
	// TODO: make sure outputs cannot escape the working directory
	// for _, o := range output {
	// 	args = append(args, "--output", o)
	// }

	if publish {
		args = append(args, "--publish")
	}

	source := dag.Directory().
		WithFile("cassette.tape", tape)

	// This is necessary due to the way Diff works with directories
	//
	// Otherwise we get an error:
	// cannot diff with different relative paths: "/" != "/work"
	sourceRoot := dag.Directory().WithDirectory("work", source)
	source = sourceRoot.Directory("work")

	result := m.Container.
		WithWorkdir("/work").
		WithMountedDirectory(".", source).
		WithExec(args).
		Directory(".")

	// Diffing is necessary because there is no way to control the output directory
	// https://github.com/charmbracelet/vhs/issues/121
	return source.Diff(result)
}

// Mount a source directory. Useful when you have source commands in your tape files.
func (m *Vhs) WithSource(
	// Source directory to mount.
	source *dagger.Directory,
) *WithSource {
	return &WithSource{
		Source: source,
		Vhs:    m,
	}
}

type WithSource struct {
	Source *dagger.Directory

	// +private
	Vhs *Vhs
}

// Runs a given tape file and generates its outputs.
func (m *WithSource) Render(
	// Tape file to render. Must be relative to the source directory.
	tape string,

	// Publish your GIF to vhs.charm.sh and get a shareable URL.
	//
	// +optional
	// +default=false
	publish bool,
) *dagger.Directory {
	// TODO: make sure tape cannot escape the working directory
	args := []string{"vhs", tape}

	if publish {
		args = append(args, "--publish")
	}

	// This is necessary due to the way Diff works with directories
	//
	// Otherwise we get an error:
	// cannot diff with different relative paths: "/" != "/work"
	source := m.Source
	sourceRoot := dag.Directory().WithDirectory("work", source)
	source = sourceRoot.Directory("work")

	result := m.Vhs.Container.
		WithWorkdir("/work").
		WithMountedDirectory(".", source).
		WithExec(args).
		Directory(".")

	// Diffing is necessary because there is no way to control the output directory
	// https://github.com/charmbracelet/vhs/issues/121
	return source.Diff(result)
}
