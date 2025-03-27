package main

import (
	"fmt"
	"strconv"
)

type Setting string

const (
	Shell         Setting = "Shell"
	FontSize      Setting = "FontSize"
	FontFamily    Setting = "FontFamily"
	Width         Setting = "Width"
	Height        Setting = "Height"
	LetterSpacing Setting = "LetterSpacing"
	LineHeight    Setting = "LineHeight"
	TypingSpeed   Setting = "TypingSpeed"
	Theme         Setting = "Theme"
	Padding       Setting = "Padding"
	Margin        Setting = "Margin"
	MarginFill    Setting = "MarginFill"
	WindowBar     Setting = "WindowBar"
	BorderRadius  Setting = "BorderRadius"
	Framerate     Setting = "Framerate"
	PlaybackSpeed Setting = "PlaybackSpeed"
	LoopOffset    Setting = "LoopOffset"
	CursorBlink   Setting = "CursorBlink"
)

func (t Tape) set(setting Setting, value string, comment string) Tape {
	return t.commandWithComment(string(Set), []string{string(setting), value}, comment)
}

// The Set command allows you to change global aspects of the terminal,
// such as the font settings, window dimensions, and GIF output location.
//
// Setting must be administered at the top of the tape file.
// Any setting (except TypingSpeed) applied after a non-setting or non-output command will be ignored.
func (t Tape) Set() TapeSetting {
	return TapeSetting{
		Tape: t.clone(),
	}
}

type TapeSetting struct {
	// +private
	Tape Tape
}

func (s TapeSetting) set(setting Setting, value string, comment string) Tape {
	return s.Tape.set(setting, value, comment)
}

// The SetBlock command allows you to change global aspects of the terminal,
// such as the font settings, window dimensions, and GIF output location.
//
// Setting must be administered at the top of the tape file.
// Any setting (except TypingSpeed) applied after a non-setting or non-output command will be ignored.
func (t Tape) SetBlock() TapeSettingBlock {
	return TapeSettingBlock{
		Tape: t.clone(),
	}
}

type TapeSettingBlock struct {
	// +private
	Tape Tape
}

func (tsb TapeSettingBlock) set(setting Setting, value string, comment string) TapeSettingBlock {
	tsb.Tape = tsb.Tape.set(setting, value, comment)

	return tsb
}

// Append a single or multiline comment to the tape.
func (tsb TapeSettingBlock) Comment(comment string) TapeSettingBlock {
	tsb.Tape = tsb.Tape.Comment(comment)

	return tsb
}

// Append a single empty line to the tape. Useful for separating commands.
func (tsb TapeSettingBlock) EmptyLine() TapeSettingBlock {
	tsb.Tape = tsb.Tape.EmptyLine()

	return tsb
}

// EndSet ends the set block.
func (tsb TapeSettingBlock) EndSet() Tape {
	return tsb.Tape
}

// Set the shell.
func (s TapeSetting) Shell(
	shell string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(Shell, shell, comment)
}

// Set the shell.
func (s TapeSettingBlock) Shell(
	shell string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(Shell, shell, comment)
}

// Set the font size.
func (s TapeSetting) FontSize(
	size int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(FontSize, strconv.Itoa(size), comment)
}

// Set the font size.
func (s TapeSettingBlock) FontSize(
	size int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(FontSize, strconv.Itoa(size), comment)
}

// Set the font family.
func (s TapeSetting) FontFamily(
	font string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(FontFamily, fmt.Sprintf("%q", font), comment)
}

// Set the font family.
func (s TapeSettingBlock) FontFamily(
	font string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(FontFamily, fmt.Sprintf("%q", font), comment)
}

// Set the width of the terminal.
func (s TapeSetting) Width(
	width int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(Width, strconv.Itoa(width), comment)
}

// Set the width of the terminal.
func (s TapeSettingBlock) Width(
	width int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(Width, strconv.Itoa(width), comment)
}

// Set the height of the terminal.
func (s TapeSetting) Height(
	height int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(Height, strconv.Itoa(height), comment)
}

// Set the height of the terminal.
func (s TapeSettingBlock) Height(
	height int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(Height, strconv.Itoa(height), comment)
}

// Set the spacing between letters (tracking).
func (s TapeSetting) LetterSpacing(
	spacing int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(LetterSpacing, strconv.Itoa(spacing), comment)
}

// Set the spacing between letters (tracking).
func (s TapeSettingBlock) LetterSpacing(
	spacing int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(LetterSpacing, strconv.Itoa(spacing), comment)
}

// Set the spacing between lines.
func (s TapeSetting) LineHeight(
	spacing float64,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(LineHeight, strconv.FormatFloat(spacing, 'f', 1, 64), comment)
}

// Set the spacing between lines.
func (s TapeSettingBlock) LineHeight(
	spacing float64,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(LineHeight, strconv.FormatFloat(spacing, 'f', 1, 64), comment)
}

// Set the typing speed of seconds per key press.
func (s TapeSetting) TypingSpeed(
	speed string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(TypingSpeed, speed, comment)
}

// Set the typing speed of seconds per key press.
func (s TapeSettingBlock) TypingSpeed(
	speed string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(TypingSpeed, speed, comment)
}

// Set the theme of the terminal.
func (s TapeSetting) Theme(
	theme string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(Theme, fmt.Sprintf("%q", theme), comment)
}

// Set the theme of the terminal.
func (s TapeSettingBlock) Theme(
	theme string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(Theme, fmt.Sprintf("%q", theme), comment)
}

// Set the padding (in pixels) of the terminal frame.
func (s TapeSetting) Padding(
	pixels int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(Padding, strconv.Itoa(pixels), comment)
}

// Set the padding (in pixels) of the terminal frame.
func (s TapeSettingBlock) Padding(
	pixels int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(Padding, strconv.Itoa(pixels), comment)
}

// Set the margin (in pixels) of the video.
func (s TapeSetting) Margin(
	pixels int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(Margin, strconv.Itoa(pixels), comment)
}

// Set the margin (in pixels) of the video.
func (s TapeSettingBlock) Margin(
	pixels int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(Margin, strconv.Itoa(pixels), comment)
}

// Set the margin fill color of the video.
func (s TapeSetting) MarginFill(
	color string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(MarginFill, fmt.Sprintf("%q", color), comment)
}

// Set the margin fill color of the video.
func (s TapeSettingBlock) MarginFill(
	color string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(MarginFill, fmt.Sprintf("%q", color), comment)
}

type WindowBarType string

const (
	Colorful      WindowBarType = "Colorful"
	ColorfulRight WindowBarType = "ColorfulRight"
	Rings         WindowBarType = "Rings"
	RingsRight    WindowBarType = "RingsRight"
)

// Set the type of window bar.
func (s TapeSetting) WindowBar(
	windowBar WindowBarType,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(WindowBar, string(windowBar), comment)
}

// Set the type of window bar.
func (s TapeSettingBlock) WindowBar(
	windowBar WindowBarType,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(WindowBar, string(windowBar), comment)
}

// Set the border radius (in pixels) of the terminal window.
func (s TapeSetting) BorderRadius(
	pixels int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(BorderRadius, strconv.Itoa(pixels), comment)
}

// Set the border radius (in pixels) of the terminal window.
func (s TapeSettingBlock) BorderRadius(
	pixels int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(BorderRadius, strconv.Itoa(pixels), comment)
}

// Set the rate at which VHS captures frames.
func (s TapeSetting) Framerate(
	rate int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(Framerate, strconv.Itoa(rate), comment)
}

// Set the rate at which VHS captures frames.
func (s TapeSettingBlock) Framerate(
	rate int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(Framerate, strconv.Itoa(rate), comment)
}

// Set the playback speed of the final render.
func (s TapeSetting) PlaybackSpeed(
	speed float64,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(PlaybackSpeed, strconv.FormatFloat(speed, 'f', 1, 64), comment)
}

// Set the playback speed of the final render.
func (s TapeSettingBlock) PlaybackSpeed(
	speed float64,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(PlaybackSpeed, strconv.FormatFloat(speed, 'f', 1, 64), comment)
}

// Set the offset for when the GIF loop should begin.
// This allows you to make the first frame of the GIF (generally used for previews) more interesting..
func (s TapeSetting) LoopOffset(
	offset string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(LoopOffset, offset, comment)
}

// Set the offset for when the GIF loop should begin.
// This allows you to make the first frame of the GIF (generally used for previews) more interesting..
func (s TapeSettingBlock) LoopOffset(
	offset string,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(LoopOffset, offset, comment)
}

// Set whether the cursor should blink. Enabled by default.
func (s TapeSetting) CursorBlink(
	blink bool,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return s.set(CursorBlink, strconv.FormatBool(blink), comment)
}

// Set whether the cursor should blink. Enabled by default.
func (s TapeSettingBlock) CursorBlink(
	blink bool,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) TapeSettingBlock {
	return s.set(CursorBlink, strconv.FormatBool(blink), comment)
}
