package main

import (
	"context"
	"dagger/helm-docs/tests/internal/dagger"
	"fmt"

	"github.com/sourcegraph/conc/pool"
)

const version = "v1.13.1"

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Default)
	p.Go(m.SortValues)
	p.Go(m.Template)

	return p.Wait()
}

func (m *Tests) Default(ctx context.Context) error {
	return m.runTest(ctx, "default", dagger.HelmDocsGenerateOpts{})
}

func (m *Tests) SortValues(ctx context.Context) error {
	return m.runTest(ctx, "sort-values", dagger.HelmDocsGenerateOpts{
		SortValuesOrder: "file",
	})
}

func (m *Tests) Template(ctx context.Context) error {
	return m.runTest(ctx, "template", dagger.HelmDocsGenerateOpts{
		Templates: []*dagger.File{
			dag.CurrentModule().Source().File("./testdata/charts/template/template.md"),
		},
	})
}

func (m *Tests) runTest(ctx context.Context, chartName string, opts dagger.HelmDocsGenerateOpts) error {
	chartDir := func(chartName string) *dagger.Directory {
		return dag.CurrentModule().Source().Directory(fmt.Sprintf("./testdata/charts/%s", chartName))
	}

	expected := func(chartName string) *dagger.File {
		return dag.CurrentModule().Source().File(fmt.Sprintf("./testdata/charts/%s/expected.md", chartName))
	}
	actual := dag.HelmDocs(dagger.HelmDocsOpts{
		Version: version,
	}).
		Generate(chartDir(chartName), opts)

	_, err := dag.Container().
		From("alpine").
		WithMountedFile("/src/expected", expected(chartName)).
		WithMountedFile("/src/actual", actual).
		WithExec([]string{"diff", "-u", "/src/expected", "/src/actual"}).
		Sync(ctx)

	return err
}
