package main

import (
	"context"
	"dagger/rust/tests/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Build)
	p.Go(m.Test)
	p.Go(m.Check)
	p.Go(m.Format)
	p.Go(m.FormatCheck)
	p.Go(m.Clippy)
	p.Go(m.ClippyFix)

	return p.Wait()
}

func (m *Tests) module() *dagger.Rust {
	return dag.Rust(dag.CurrentModule().Source().Directory("testdata"))
}

func (m *Tests) Build(ctx context.Context) error {
	rust := m.module()

	_, err := rust.Build().Sync(ctx)

	return err
}

func (m *Tests) Test(ctx context.Context) error {
	rust := m.module()

	_, err := rust.Test().Sync(ctx)

	return err
}

func (m *Tests) Check(ctx context.Context) error {
	rust := m.module()

	_, err := rust.Check().Sync(ctx)

	return err
}

func (m *Tests) Format(ctx context.Context) error {
	rust := m.module()

	_, err := rust.Format().Run().Sync(ctx)

	return err
}

func (m *Tests) FormatCheck(ctx context.Context) error {
	rust := m.module()

	_, err := rust.Format().Check().Sync(ctx)

	return err
}

func (m *Tests) Clippy(ctx context.Context) error {
	rust := m.module()

	_, err := rust.Clippy().Run().Sync(ctx)

	return err
}

func (m *Tests) ClippyFix(ctx context.Context) error {
	rust := m.module()

	_, err := rust.Clippy().Fix().Sync(ctx)

	return err
}
