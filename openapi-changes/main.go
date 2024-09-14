// The world's sexiest OpenAPI breaking changes detector.
//
// Discover what changed between two OpenAPI specs, or a single spec over time.
//
// Supports OpenAPI 3.1, 3.0 and Swagger

package main

import (
	"context"
	"dagger/openapi-changes/internal/dagger"
	"fmt"
	"strings"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "pb33f/openapi-changes"

type OpenapiChanges struct {
	// +private
	Container *dagger.Container

	// +private
	NoStyle bool
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

	// Disable all color output and all terminal styling (useful for CI/CD).
	//
	// +optional
	noStyle bool,
) *OpenapiChanges {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	return &OpenapiChanges{
		Container: container,
		NoStyle:   noStyle,
	}
}

// Check for changes in a local Git repository.
func (m *OpenapiChanges) Local(
	// The Git repository containing the OpenAPI specs.
	source *dagger.Directory,

	// Path to the OpenAPI spec file.
	spec string,

	// Disable all color output and all terminal styling (useful for CI/CD).
	//
	// +optional
	noStyle bool,

	// Only show the latest changes (the last git revision against HEAD).
	//
	// +optional
	top bool,

	// Limit the number of changes to show when using git.
	//
	// +optional
	limit int,

	// Base URL or Base working directory to use for relative references.
	//
	// +optional
	base string,
) *Run {
	cmdArgs := cmdArgs{
		noStyle: m.NoStyle || noStyle,
		top:     top,
		limit:   limit,
		base:    base,
	}

	return &Run{
		Container: m.Container.
			WithMountedDirectory("/work", source).
			WithWorkdir("/work"),
		Args: cmdArgs.append([]string{".", spec}),
	}
}

// Check for changes in a remote Git repository.
// (Currently only supports GitHub)
func (m *OpenapiChanges) Remote(
	// The URL of the OpenAPI spec file.
	url string,

	// Disable all color output and all terminal styling (useful for CI/CD).
	//
	// +optional
	noStyle bool,

	// Only show the latest changes (the last git revision against HEAD).
	//
	// +optional
	top bool,

	// Limit the number of changes to show when using git.
	//
	// +optional
	limit int,

	// Base URL or Base working directory to use for relative references.
	//
	// +optional
	base string,
) *Run {
	cmdArgs := cmdArgs{
		noStyle: m.NoStyle || noStyle,
		top:     top,
		limit:   limit,
		base:    base,
	}

	return &Run{
		Container: m.Container.
			WithWorkdir("/work"),
		Args: cmdArgs.append([]string{url}),
	}
}

// Compare two OpenAPI specs.
func (m *OpenapiChanges) Diff(
	// The old OpenAPI spec file.
	old *dagger.File,

	// The new OpenAPI spec file.
	new *dagger.File,

	// Optional source directory to use for relative references.
	//
	// +optional
	source *dagger.Directory,

	// Disable all color output and all terminal styling (useful for CI/CD).
	//
	// +optional
	noStyle bool,

	// Base URL or Base working directory to use for relative references.
	//
	// +optional
	base string,
) *Run {
	cmdArgs := cmdArgs{
		noStyle: m.NoStyle || noStyle,
		base:    base,
	}

	return &Run{
		Container: m.Container.
			WithMountedFile("/work/old.yaml", old).
			WithMountedFile("/work/new.yaml", new).
			With(func(c *dagger.Container) *dagger.Container {
				if source != nil {
					c = c.WithMountedDirectory("/work/src", source)
				}

				return c
			}).
			WithWorkdir("/work/src"),
		Args: cmdArgs.append([]string{"/work/old.yaml", "/work/new.yaml"}),
	}
}

type Run struct {
	// +private
	Container *dagger.Container

	// +private
	Args []string
}

func (r *Run) Summary(ctx context.Context) (string, error) {
	args := append([]string{"openapi-changes", "summary"}, r.Args...)

	return r.Container.WithExec([]string{"sh", "-c", strings.Join(args, " ") + " || exit 0"}).Stdout(ctx)
}

func (r *Run) Markdown(ctx context.Context) (*dagger.File, error) {
	args := append([]string{"openapi-changes", "summary", "--markdown"}, r.Args...)

	output, err := r.Container.WithExec([]string{"sh", "-c", strings.Join(args, " ") + " || exit 0"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	return dag.Directory().WithNewFile("summary.md", output).File("summary.md"), nil
}

func (r *Run) Json(ctx context.Context) (*dagger.File, error) {
	args := append([]string{"openapi-changes", "report"}, r.Args...)

	output, err := r.Container.WithExec([]string{"sh", "-c", strings.Join(args, " ") + " || exit 0"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	return dag.Directory().WithNewFile("report.json", output).File("report.json"), nil
}

func (r *Run) HTML() *dagger.File {
	args := append([]string{"openapi-changes", "html-report"}, r.Args...)

	return r.Container.WithExec([]string{"sh", "-c", strings.Join(args, " ") + " || exit 0"}).File("report.html")
}

type cmdArgs struct {
	// Disable all color output and all terminal styling (useful for CI/CD).
	noStyle bool

	// Only show the latest changes (the last git revision against HEAD).
	top bool

	// Limit the number of changes to show when using git.
	limit int

	// Base URL or Base working directory to use for relative references.
	base string
}

func (a cmdArgs) append(args []string) []string {
	if a.noStyle {
		args = append(args, "--no-style")
	}

	if a.top {
		args = append(args, "--top")
	}

	if a.limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", a.limit))
	}

	if a.base != "" {
		args = append(args, "--base", fmt.Sprintf("%s", a.base))
	}

	return args
}
