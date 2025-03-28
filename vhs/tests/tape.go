package main

import (
	"context"
	"dagger/vhs/tests/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

// Tape tests
func (m *Tests) Tape() *Tape {
	return &Tape{}
}

type Tape struct{}

// All executes all tests.
func (m *Tape) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

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

func testTape(ctx context.Context, tape *dagger.VhsTape, expected string) error {
	_, err := dag.Container().
		From("alpine:latest").
		WithWorkdir("/work").
		WithMountedFile("expected.tape", dag.CurrentModule().Source().File("testdata/"+expected)).
		WithMountedFile("actual.tape", tape.EmptyLine().File()). // Most editors append a trailing new line to files
		WithExec([]string{"diff", "expected.tape", "actual.tape"}).
		Sync(ctx)

	return err
}

func (m *Tape) Output(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Output("out.gif").
		Output("out.mp4").
		Output("out.webm").
		Output("frames/", dagger.VhsTapeOutputOpts{
			Comment: "a directory of frames as a PNG sequence",
		})

	return testTape(ctx, tape, "output.tape")
}

func (m *Tape) Require(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Comment("A tape file that requires gum and glow to be in the $PATH").
		Require("gum").
		Require("glow")

	return testTape(ctx, tape, "require.tape")
}

func (m *Tape) Set(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Set().Shell("fish").
		EmptyLine().
		Set().FontSize(10).
		Set().FontSize(20).
		Set().FontSize(40).
		EmptyLine().
		Set().FontFamily("Monoflow").
		EmptyLine().
		Set().Width(300).
		EmptyLine().
		Set().Height(1000).
		EmptyLine().
		Set().LetterSpacing(20).
		EmptyLine().
		Set().LineHeight(1.8).
		EmptyLine().
		Set().TypingSpeed("500ms", dagger.VhsTapeSettingTypingSpeedOpts{Comment: "500ms"}).
		Set().TypingSpeed("1s", dagger.VhsTapeSettingTypingSpeedOpts{Comment: "1s"}).
		Set().TypingSpeed("0.1").
		EmptyLine().
		Set().Theme("Catppuccin Frappe").
		EmptyLine().
		Set().Padding(0).
		EmptyLine().
		Set().Margin(60).
		Set().MarginFill("#6B50FF").
		EmptyLine().
		Set().WindowBar(dagger.VhsWindowBarTypeColorful).
		EmptyLine().
		Comment("You'll likely want to add a Margin + MarginFill if you use BorderRadius.").
		Set().Margin(20).
		Set().MarginFill("#674EFF").
		Set().BorderRadius(10).
		EmptyLine().
		Set().Framerate(60).
		EmptyLine().
		Set().PlaybackSpeed(0.5, dagger.VhsTapeSettingPlaybackSpeedOpts{Comment: "Make output 2 times slower"}).
		Set().PlaybackSpeed(1.0, dagger.VhsTapeSettingPlaybackSpeedOpts{Comment: "Keep output at normal speed (default)"}).
		Set().PlaybackSpeed(2.0, dagger.VhsTapeSettingPlaybackSpeedOpts{Comment: "Make output 2 times faster"}).
		EmptyLine().
		Set().LoopOffset("5", dagger.VhsTapeSettingLoopOffsetOpts{Comment: "Start the GIF at the 5th frame"}).
		Set().LoopOffset("50%", dagger.VhsTapeSettingLoopOffsetOpts{Comment: "Start the GIF halfway through"}).
		EmptyLine().
		Set().CursorBlink(false)

	return testTape(ctx, tape, "set.tape")
}

func (m *Tape) SetBlock(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		SetBlock().
		Shell("fish").
		EmptyLine().
		FontSize(10).
		FontSize(20).
		FontSize(40).
		EmptyLine().
		FontFamily("Monoflow").
		EmptyLine().
		Width(300).
		EmptyLine().
		Height(1000).
		EmptyLine().
		LetterSpacing(20).
		EmptyLine().
		LineHeight(1.8).
		EmptyLine().
		TypingSpeed("500ms", dagger.VhsTapeSettingBlockTypingSpeedOpts{Comment: "500ms"}).
		TypingSpeed("1s", dagger.VhsTapeSettingBlockTypingSpeedOpts{Comment: "1s"}).
		TypingSpeed("0.1").
		EmptyLine().
		Theme("Catppuccin Frappe").
		EmptyLine().
		Padding(0).
		EmptyLine().
		Margin(60).
		MarginFill("#6B50FF").
		EmptyLine().
		WindowBar(dagger.VhsWindowBarTypeColorful).
		EmptyLine().
		Comment("You'll likely want to add a Margin + MarginFill if you use BorderRadius.").
		Margin(20).
		MarginFill("#674EFF").
		BorderRadius(10).
		EmptyLine().
		Framerate(60).
		EmptyLine().
		PlaybackSpeed(0.5, dagger.VhsTapeSettingBlockPlaybackSpeedOpts{Comment: "Make output 2 times slower"}).
		PlaybackSpeed(1.0, dagger.VhsTapeSettingBlockPlaybackSpeedOpts{Comment: "Keep output at normal speed (default)"}).
		PlaybackSpeed(2.0, dagger.VhsTapeSettingBlockPlaybackSpeedOpts{Comment: "Make output 2 times faster"}).
		EmptyLine().
		LoopOffset("5", dagger.VhsTapeSettingBlockLoopOffsetOpts{Comment: "Start the GIF at the 5th frame"}).
		LoopOffset("50%", dagger.VhsTapeSettingBlockLoopOffsetOpts{Comment: "Start the GIF halfway through"}).
		EmptyLine().
		CursorBlink(false).
		EndSet()

	return testTape(ctx, tape, "set.tape")
}

func (m *Tape) Type(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Comment("Type something").
		Type("Whatever you want").
		EmptyLine().
		Comment("Type something really slowly!").
		Type("Slow down there, partner.", dagger.VhsTapeTypeOpts{Time: "500ms"}).
		EmptyLine().
		Type(`VAR="Escaped"`)

	return testTape(ctx, tape, "type.tape")
}

func (m *Tape) Keys(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Backspace(dagger.VhsTapeBackspaceOpts{Count: 18}).
		EmptyLine().
		Ctrl("R").
		EmptyLine().
		Enter(dagger.VhsTapeEnterOpts{Count: 2}).
		EmptyLine().
		Up(dagger.VhsTapeUpOpts{Count: 2}).
		Down(dagger.VhsTapeDownOpts{Count: 2}).
		Left().
		Right().
		Left().
		Right().
		Type("B").
		Type("A").
		EmptyLine().
		Tab(dagger.VhsTapeTabOpts{Time: "500ms", Count: 2}).
		EmptyLine().
		Space(dagger.VhsTapeSpaceOpts{Count: 10}).
		EmptyLine().
		PageUp(dagger.VhsTapePageUpOpts{Count: 3}).
		PageDown(dagger.VhsTapePageDownOpts{Count: 5})

	return testTape(ctx, tape, "keys.tape")
}

func (m *Tape) Wait(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Wait().
		Wait(dagger.VhsTapeWaitOpts{Regexp: "World"}).
		Wait(dagger.VhsTapeWaitOpts{Scope: dagger.VhsWaitScopeScreen, Regexp: "World"}).
		Wait(dagger.VhsTapeWaitOpts{Scope: dagger.VhsWaitScopeLine, Regexp: "World"}).
		Wait(dagger.VhsTapeWaitOpts{Timeout: "10ms", Regexp: "World"}).
		Wait(dagger.VhsTapeWaitOpts{Scope: dagger.VhsWaitScopeLine, Timeout: "10ms", Regexp: "World"})

	return testTape(ctx, tape, "wait.tape")
}

func (m *Tape) Sleep(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Sleep("0.5", dagger.VhsTapeSleepOpts{Comment: "500ms"}).
		Sleep("2", dagger.VhsTapeSleepOpts{Comment: "2s"}).
		Sleep("100ms", dagger.VhsTapeSleepOpts{Comment: "100ms"}).
		Sleep("1s", dagger.VhsTapeSleepOpts{Comment: "1s"})

	return testTape(ctx, tape, "sleep.tape")
}

func (m *Tape) ShowHide(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Hide().
		Type("You won't see this being typed.").
		Show().
		Type("You will see this being typed.")

	return testTape(ctx, tape, "show-hide.tape")
}

func (m *Tape) Screenshot(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Comment("At any point...").
		Screenshot("examples/screenshot.png")

	return testTape(ctx, tape, "screenshot.tape")
}

func (m *Tape) CopyPaste(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Copy("https://github.com/charmbracelet").
		Type("open ").
		Sleep("500ms").
		Paste()

	return testTape(ctx, tape, "copy-paste.tape")
}

func (m *Tape) Env(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Env("HELLO", "WORLD").
		EmptyLine().
		Type("echo $HELLO").
		Enter().
		Sleep("1s")

	return testTape(ctx, tape, "env.tape")
}

func (m *Tape) Source(ctx context.Context) error {
	vhs := dag.Vhs()

	tape := vhs.Edit().
		Source("config.tape")

	return testTape(ctx, tape, "source.tape")
}
