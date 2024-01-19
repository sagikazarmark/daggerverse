package main

import (
	"context"
	"fmt"
	"path"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "stoplight/spectral"

type Spectral struct {
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
) *Spectral {
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

	return &Spectral{ctr}
}

func (m *Spectral) Container() *Container {
	return m.Ctr
}

// Lint JSON/YAML documents.
func (m *Spectral) Lint(
	ctx context.Context,

	// JSON/YAML OpenAPI documents.
	documents []*File,

	// Ruleset file.
	ruleset *File,

	// Results of this level or above will trigger a failure exit code. (choices: "error", "warn", "info", "hint") (default "error")
	// +optional
	failSeverity string,

	// Only output results equal to or greater than fail severity.
	// +optional
	displayOnlyFailures bool,

	// Custom json-ref-resolver instance.
	// +optional
	resolver *File,

	// Text encoding to use. (choices: "utf8", "ascii", "utf-8", "utf16le", "ucs2", "ucs-2", "base64", "latin1") (default "utf8")
	// +optional
	encoding string,

	// Increase verbosity.
	// +optional
	verbose bool,

	// No logging, output only.
	// +optional
	quiet bool,
) (*Container, error) {
	ctr := m.Ctr
	args := []string{"lint"}

	{
		dir := dag.Directory().WithFile("", ruleset)

		entries, err := dir.Entries(ctx)
		if err != nil {
			return nil, err
		}

		if len(entries) < 1 {
			return nil, fmt.Errorf("ruleset file is missing")
		}

		ctr = ctr.WithMountedDirectory("/work/ruleset", dir)
		args = append(args, "--ruleset", path.Join("/work/ruleset", entries[0]))
	}

	if failSeverity != "" {
		args = append(args, "--fail-severity", failSeverity)
	}

	if resolver != nil {
		dir := dag.Directory().WithFile("", resolver)

		entries, err := dir.Entries(ctx)
		if err != nil {
			return nil, err
		}

		if len(entries) < 1 {
			return nil, fmt.Errorf("resolver file is missing")
		}

		ctr = ctr.WithMountedDirectory("/work/resolver", dir)
		args = append(args, "--resolver", path.Join("/work/resolver", entries[0]))
	}

	if verbose {
		args = append(args, "--verbose")
	}

	if quiet {
		args = append(args, "--quiet")
	}

	{
		dir := dag.Directory()

		for _, document := range documents {
			dir = dir.WithFile("", document)
		}

		entries, err := dir.Entries(ctx)
		if err != nil {
			return nil, err
		}

		ctr = ctr.WithMountedDirectory("/work/documents", dir)

		for _, e := range entries {
			args = append(args, path.Join("/work/documents", e))
		}
	}

	return ctr.WithExec(args), nil
}
