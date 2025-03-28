package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(
	ctx context.Context,

	// Do not run tape tests (sometimes they are slow)
	//
	// +optional
	// +default=false
	withoutTape bool,
) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Render)
	p.Go(m.Render_Advanced)
	p.Go(m.WithSource_Render)
	p.Go(m.WithSource_Render_Advanced)

	if !withoutTape {
		p.Go(m.Tape().All)
	}

	return p.Wait()
}

func (m *Tests) Render(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.NewTape()

	entries, err := vhs.Render(tape).Entries(ctx)
	if err != nil {
		return err
	}

	if !slices.Equal(entries, []string{"cassette.gif"}) {
		return fmt.Errorf("unexpected entries: %v", entries)
	}

	return nil
}

func (m *Tests) Render_Advanced(ctx context.Context) error {
	vhs := dag.Vhs()

	out := vhs.Edit().
		Output("cassette.webm").
		Output("dir/cassette.gif").
		Type("echo Hello").
		Enter().
		Render()

	{
		entries, err := out.Entries(ctx)
		if err != nil {
			return err
		}

		if !slices.Equal(entries, []string{"cassette.webm", "dir/"}) {
			return fmt.Errorf("unexpected entries: %v", entries)
		}
	}

	{
		entries, err := out.Directory("dir").Entries(ctx)
		if err != nil {
			return err
		}

		if !slices.Equal(entries, []string{"cassette.gif"}) {
			return fmt.Errorf("unexpected entries: %v", entries)
		}
	}

	return nil
}

func (m *Tests) WithSource_Render(ctx context.Context) error {
	vhs := dag.Vhs()

	dir := dag.Directory().WithFile("cassette.tape", vhs.NewTape())

	entries, err := vhs.WithSource(dir).Render("cassette.tape").Entries(ctx)
	if err != nil {
		return err
	}

	if !slices.Equal(entries, []string{"cassette.gif"}) {
		return fmt.Errorf("unexpected entries: %v", entries)
	}

	return nil
}

func (m *Tests) WithSource_Render_Advanced(ctx context.Context) error {
	vhs := dag.Vhs()

	tape1 := vhs.Edit().
		Output("cassette.gif").
		Output("dir/cassette.gif").
		Type("echo Hello").
		Enter().
		File()

	tape2 := vhs.Edit().
		Output("cassette.webm").
		Output("dir/cassette.gif").
		Type("echo Hello").
		Enter().
		File()

	dir := dag.Directory().
		WithFile("cassette1.tape", tape1).
		WithFile("dir/cassette2.tape", tape2)

	out := vhs.WithSource(dir).Render("dir/cassette2.tape")

	{
		entries, err := out.Entries(ctx)
		if err != nil {
			return err
		}

		if !slices.Equal(entries, []string{"cassette.webm", "dir/"}) {
			return fmt.Errorf("unexpected entries: %v", entries)
		}
	}

	{
		entries, err := out.Directory("dir").Entries(ctx)
		if err != nil {
			return err
		}

		if !slices.Equal(entries, []string{"cassette.gif"}) {
			return fmt.Errorf("unexpected entries: %v", entries)
		}
	}

	return nil
}
