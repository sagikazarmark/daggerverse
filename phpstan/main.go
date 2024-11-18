// PHP Static Analysis Tool - discover bugs in your code without running it!

package main

import (
	"dagger/phpstan/internal/dagger"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "ghcr.io/phpstan/phpstan"

type Phpstan struct {
	Container *dagger.Container
}

// TODO: add support for custom cache

func New(
	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	// +default="latest"
	version string,

	// Customize PHP version (currently supported: any minor version from the 8.x branch).
	//
	// +optional
	phpVersion string,

	// Custom container to use as a base container. Takes precedence over version.
	//
	// +optional
	container *dagger.Container,
) *Phpstan {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		if phpVersion != "" {
			version = version + "-php" + phpVersion
		}

		container = dag.Container().From(defaultImageRepository + ":" + version)
	}

	return &Phpstan{
		Container: container,
	}
}

// TODO: add support for custom config

// Analyse source code.
func (m *Phpstan) Analyse(
	source *dagger.Directory,

	// Paths with source code to run analysis on.
	//
	// +optional
	paths []string,
) *dagger.Container {
	args := []string{"phpstan", "analyse"}

	if len(paths) > 0 {
		args = append(args, paths...)
	}

	return m.Container.
		WithWorkdir("/work/source").
		WithMountedDirectory("/work/source", source).
		WithExec(args)
}

// Generate baseline file.
func (m *Phpstan) GenerateBaseline(
	source *dagger.Directory,

	// Paths with source code to run analysis on.
	//
	// +optional
	paths []string,
) *dagger.File {
	args := []string{"phpstan", "analyse", "--generate-baseline"}

	if len(paths) > 0 {
		args = append(args, paths...)
	}

	return m.Container.
		WithWorkdir("/work/source").
		WithMountedDirectory("/work/source", source).
		WithExec(args).
		File("phpstan-baseline.neon")
}
