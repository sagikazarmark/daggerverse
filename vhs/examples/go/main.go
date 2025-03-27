package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Examples struct{}

// All executes all examples.
func (m *Examples) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.VhsRender)
	// p.Go(m.VhsRender_Output)
	p.Go(m.VhsEdit)
	p.Go(m.VhsWithSource)

	return p.Wait()
}

func (m *Examples) VhsRender(ctx context.Context) error {
	vhs := dag.Vhs()

	// Create a new tape (or load an existing one)
	tape := vhs.NewTape()

	out, err := vhs.Render(tape).Sync(ctx)
	if err != nil {
		return err
	}

	// The output is a directory containing the rendered files.
	_ = out

	return nil
}

// Does not seem to work at the moment
// func (m *Examples) VhsRender_Output(ctx context.Context) error {
// 	vhs := dag.Vhs()

// 	// Create a new tape (or load an existing one)
// 	tape := vhs.NewTape()

// 	_, err := vhs.Render(tape, dagger.VhsRenderOpts{
// 		// Customize output options
// 		Output: []string{"cassette.ascii"},
// 	}).File("cassette.ascii").Contents(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (m *Examples) VhsEdit(ctx context.Context) error {
	vhs := dag.Vhs()

	// Create a new tape
	tape := vhs.Edit().
		Comment("Hello world").
		EmptyLine().

		// Set some outputs
		Output("out.gif").
		Output("out.webm").
		EmptyLine().

		// Set some settings
		Set().FontSize(14).
		Set().FontFamily("Monoflow").
		EmptyLine().

		// Use setBlock for more than one settings for brevity
		SetBlock().
		FontSize(16).
		FontFamily("Iosevka").
		EndSet().
		EmptyLine().

		// Do something
		Type("echo Hello world").
		Enter().
		Sleep("1s")

	// Get the tape file
	_, err := tape.File().Sync(ctx)
	if err != nil {
		return err
	}

	// Get the outputs
	_, err = tape.Render().Sync(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m *Examples) VhsWithSource(ctx context.Context) error {
	vhs := dag.Vhs()

	// Create some tapes
	config := vhs.Edit().
		SetBlock().
		FontSize(16).
		FontFamily("Iosevka").
		EndSet()

	tape := vhs.Edit().
		Source("config.tape").
		Type("echo Hello world").
		Enter()

	tapes := dag.Directory().
		WithFile("config.tape", config.File()).
		WithFile("cassette.tape", tape.File())

	_, err := vhs.WithSource(tapes).Render("cassette.tape").Sync(ctx)
	if err != nil {
		return err
	}

	return nil
}
