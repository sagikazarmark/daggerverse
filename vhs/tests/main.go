package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Render)

	// Tape
	p.Go(m.Output)
	p.Go(m.Require)
	p.Go(m.Set)
	p.Go(m.SetBlock)
	p.Go(m.Type)
	p.Go(m.Keys)
	p.Go(m.Wait)
	p.Go(m.Sleep)
	p.Go(m.ShowHide)
	p.Go(m.Screenshot)
	p.Go(m.CopyPaste)
	p.Go(m.Env)
	p.Go(m.Source)

	return p.Wait()
}

func (m *Tests) Render(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.NewTape()

	_, err := vhs.Render(tape).Sync(ctx)

	return err
}
