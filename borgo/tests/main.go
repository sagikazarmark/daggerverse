package main

import (
	"context"
	"fmt"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Compile)

	return p.Wait()
}

func (m *Tests) Compile(ctx context.Context) error {
	source := dag.CurrentModule().Source().Directory("testdata/hello-world")
	source = dag.Borgo().Compile(source)

	helloWorld := dag.Go().Build(source)

	const expected = "Hello world\n"

	actual, err := dag.Container().
		From("alpine").
		WithFile("/hello-world", helloWorld).
		WithExec([]string{"/hello-world"}).
		Stdout(ctx)
	if err != nil {
		return err
	}

	if actual != expected {
		return fmt.Errorf("expected %q, got %q", expected, actual)
	}

	return nil
}
