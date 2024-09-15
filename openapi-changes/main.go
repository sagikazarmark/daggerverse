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

// Return a summary of the changes.
//
// See https://pb33f.io/openapi-changes/summary/ for more information.
func (r *Run) Summary(ctx context.Context) (string, error) {
	args := append([]string{"openapi-changes", "summary"}, r.Args...)

	return r.Container.WithExec([]string{"sh", "-c", strings.Join(args, " ") + " || exit 0"}).Stdout(ctx)
}

// Return a summary of the changes in Markdown format.
//
// See https://pb33f.io/openapi-changes/summary/ for more information.
func (r *Run) Markdown(ctx context.Context) (*dagger.File, error) {
	args := append([]string{"openapi-changes", "summary", "--markdown"}, r.Args...)

	output, err := r.Container.WithExec([]string{"sh", "-c", strings.Join(args, " ") + " || exit 0"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	return dag.Directory().WithNewFile("summary.md", output).File("summary.md"), nil
}

// Return a JSON report of the changes.
//
// See https://pb33f.io/openapi-changes/report/ for more information.
func (r *Run) Json(ctx context.Context) (*dagger.File, error) {
	args := append([]string{"openapi-changes", "report"}, r.Args...)

	output, err := r.Container.WithExec([]string{"sh", "-c", strings.Join(args, " ") + " || exit 0"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	return dag.Directory().WithNewFile("report.json", output).File("report.json"), nil
}

// Return a HTML report of the changes.
//
// See https://pb33f.io/openapi-changes/html-report/ for more information.
func (r *Run) HTML() *HtmlReport {
	args := append([]string{"openapi-changes", "html-report"}, r.Args...)

	return &HtmlReport{
		File: r.Container.WithExec([]string{"sh", "-c", strings.Join(args, " ") + " || exit 0"}).File("report.html"),
	}
}

type HtmlReport struct {
	File *dagger.File
}

// Serve the HTML report on a local server.
func (r *HtmlReport) Serve(
	// The port to serve the HTML report on.
	//
	// +optional
	// +default=8080
	port int,
) *dagger.Service {
	return dag.Container().
		From("caddy:2-alpine").
		WithExposedPort(port).
		WithMountedFile("/var/www/index.html", r.File).
		WithWorkdir("/var/www").
		WithExec([]string{"caddy", "file-server", "--listen", fmt.Sprintf(":%d", port)}).
		AsService()
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
