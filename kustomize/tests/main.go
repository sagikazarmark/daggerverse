package main

import (
	"context"
	"errors"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Build)

	return p.Wait()
}

func (m *Tests) Build(ctx context.Context) error {
	expected, err := dag.CurrentModule().Source().File("./testdata/sample/output.yaml").Contents(ctx)
	if err != nil {
		return err
	}

	actual, err := dag.Kustomize().Build(dag.CurrentModule().Source().Directory("./testdata/sample")).Contents(ctx)
	if err != nil {
		return err
	}

	if expected != actual {
		return errors.New("expected and actual output do not match")
	}

	return nil
}

func (m *Tests) Edit(ctx context.Context) error {
	source := dag.CurrentModule().Source().Directory("./testdata/sample")

	source = dag.Kustomize().
		Edit(source).
		Set().Annotation("foo", "bar").
		Set().Image("nginx:1.16").
		Set().Namespace("non-default").
		Directory()

	expected, err := dag.CurrentModule().Source().File("./testdata/outputs/kustomization.edit.yaml").Contents(ctx)
	if err != nil {
		return err
	}

	actual, err := source.File("kustomization.yaml").Contents(ctx)
	if err != nil {
		return err
	}

	if expected != actual {
		return errors.New("expected and actual output do not match")
	}

	return nil
}
