package main

import (
	"dagger/vhs/internal/dagger"
	"fmt"
	"slices"
	"strings"
)

// Create a new tape manually.
func (m *Vhs) Tape() Tape {
	return Tape{
		Vhs: m,
	}
}

type Tape struct {
	// +private
	Entries []string

	// +private
	Vhs *Vhs
}

// Get the final tape file.
func (t Tape) File(
	// Name of the tape file to create.
	//
	// +optional
	// +default="cassette.tape"
	name string,
) *dagger.File {
	if name == "" {
		name = "cassette.tape"
	}

	contents := strings.Join(t.Entries, "\n")

	return dag.Directory().WithNewFile(name, contents).File(name)
}

// Runs the tape file and generates its outputs.
//
// Do not use source commands in your tape file.
func (t Tape) Render(
	// Publish your GIF to vhs.charm.sh and get a shareable URL.
	//
	// +optional
	// +default=false
	publish bool,
) *dagger.Directory {
	return t.Vhs.Render(t.File(""), publish)
}

func (t Tape) clone() Tape {
	t.Entries = slices.Clone(t.Entries)

	return t
}

func (t *Tape) append(lines ...string) {
	t.Entries = append(t.Entries, lines...)
}

// Append a raw line to the tape.
func (t Tape) raw(line string) Tape {
	t = t.clone()
	t.append(line)

	return t
}

// quote a string according to tape file syntax.
func quote(s string) string {
	if strings.Contains(s, `"`) {
		return fmt.Sprintf("`%s`", s)
	}

	return fmt.Sprintf("%q", s)
}

type Command string

const (
	Output  Command = "Output"
	Require Command = "Require"
	Set     Command = "Set"
	Type    Command = "Type"

	Left      Command = "Left"
	Right     Command = "Right"
	Up        Command = "Up"
	Down      Command = "Down"
	Backspace Command = "Backspace"
	Enter     Command = "Enter"
	Tab       Command = "Tab"
	Space     Command = "Space"
	Ctrl      Command = "Ctrl"
	PageUp    Command = "PageUp"
	PageDown  Command = "PageDown"

	Sleep Command = "Sleep"
	Wait  Command = "Wait"

	Hide Command = "Hide"
	Show Command = "Show"

	Screenshot Command = "Screenshot"

	Copy  Command = "Copy"
	Paste Command = "Paste"

	Source Command = "Source"
	Env    Command = "Env"
)

func (t Tape) command(command string, args []string) Tape {
	return t.commandWithComment(command, args, "")
}

func (t Tape) commandWithComment(command string, args []string, comment string) Tape {
	line := command

	if len(args) > 0 {
		line += " " + strings.Join(args, " ")
	}

	if comment != "" {
		line += " # " + strings.ReplaceAll(strings.TrimSpace(strings.TrimPrefix(comment, "#")), "\n", " ")
	}

	return t.raw(line)
}

// Append a single or multiline comment to the tape.
func (t Tape) Comment(comment string) Tape {
	// Make sure each line is prefixed with a comment symbol
	lines := strings.Split(comment, "\n")

	for i, line := range lines {
		lines[i] = fmt.Sprintf("# %s", line)
	}

	// This is a special multiline case, so we don't use [Raw]
	t = t.clone()
	t.append(lines...)

	return t
}

// Append a single empty line to the tape. Useful for separating commands.
func (t Tape) EmptyLine() Tape {
	return t.raw("")
}

// The Output command allows you to specify the location and file format of the render.
// You can specify more than one output in a tape file which will render them to the respective locations.
func (t Tape) Output(
	path string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.commandWithComment(string(Output), []string{path}, comment)
}

// The Require command allows you to specify dependencies for your tape file.
// These are useful to fail early if a required program is missing from the $PATH, and it is certain that the VHS execution will not work as expected.
//
// Require commands must be defined at the top of a tape file, before any non- setting or non-output command.
func (t Tape) Require(
	program string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.commandWithComment(string(Require), []string{program}, comment)
}

// Use Type to emulate key presses.
// That is, you can use Type to script typing in a terminal.
// Type is handy for both entering commands and interacting with prompts and TUIs in the terminal.
// The command takes a string argument of the characters to type.
func (t Tape) Type(
	characters string,

	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	command := string(Type)

	if time != "" {
		command += "@" + time
	}

	return t.commandWithComment(command, []string{quote(characters)}, comment)
}

type WaitScope string

const (
	Line   WaitScope = "Line"
	Screen WaitScope = "Screen"
)

// The Wait command allows you to wait for something to appear on the screen.
// This is useful when you need to wait on something to complete,
// even if you don't know how long it'll take,
// while including it in the recording like a spinner or loading state.
// The command takes a regular expression as an argument,
// and optionally allows to set the duration to wait
// and if you want to check the whole screen or just the last line (the scope).
func (t Tape) Wait(
	// Regular expression to wait for.
	//
	// +optional
	regexp string,

	// Scope to wait for the regular expression to appear (whole screen or just the last line).
	//
	// +optional
	scope WaitScope,

	// Duration to wait for the regular expression to appear.
	//
	// +optional
	timeout string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	command := string(Wait)

	if scope != "" {
		command += "+" + string(scope)
	}

	if timeout != "" {
		command += "@" + timeout
	}

	var args []string

	if regexp != "" {
		args = append(args, fmt.Sprintf("/%s/", regexp))
	}

	return t.commandWithComment(command, args, comment)
}

// The Sleep command allows you to continue capturing frames without interacting with the terminal.
// This is useful when you need to wait on something to complete while including it in the recording
// like a spinner or loading state. The command takes a number argument in seconds.
func (t Tape) Sleep(
	seconds string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.commandWithComment(string(Sleep), []string{seconds}, comment)
}

// The Hide command instructs VHS to stop capturing frames.
// It's useful to pause a recording to perform hidden commands.
func (t Tape) Hide(
	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.commandWithComment(string(Hide), nil, comment)
}

// The Show command instructs VHS to begin capturing frames, again.
// It's useful after a Hide command to resume frame recording for the output.
func (t Tape) Show(
	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.commandWithComment(string(Show), nil, comment)
}

// The Screenshot command captures the current frame (png format).
func (t Tape) Screenshot(
	path string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.commandWithComment(string(Screenshot), []string{path}, comment)
}

// The Copy command copies a value to the clipboard.
func (t Tape) Copy(
	value string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.commandWithComment(string(Copy), []string{quote(value)}, comment)
}

// The Paste command pastes the value from clipboard.
func (t Tape) Paste(
	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.commandWithComment(string(Paste), nil, comment)
}

// The Env command sets the environment variable via key-value pair.
func (t Tape) Env(
	key string,
	value string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.commandWithComment(string(Env), []string{key, quote(value)}, comment)
}

// The Source command allows you to execute commands from another tape.
func (t Tape) Source(
	tape string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.commandWithComment(string(Source), []string{tape}, comment)
}
