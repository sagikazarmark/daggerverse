package main

import (
	"fmt"
)

func (t Tape) key(key string, time string, count int, comment string) Tape {
	if time != "" {
		key += "@" + time
	}

	var args []string

	if count > 0 {
		args = append(args, fmt.Sprintf("%d", count))
	}

	return t.commandWithComment(key, args, comment)
}

// Press the backspace key.
func (t Tape) Backspace(
	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Repeat the key press.
	//
	// +optional
	count int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.key(string(Backspace), time, count, comment)
}

// Access the control modifier and send control sequences.
func (t Tape) Ctrl(
	char string,

	// Press the "alt" key.
	//
	// +optional
	// +default=false
	alt bool,

	// Press the "shift" key.
	//
	// +optional
	// +default=false
	shift bool,

	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Repeat the key press.
	//
	// +optional
	count int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	key := string(Ctrl)

	if alt {
		key += "+Alt"
	}

	if shift {
		key += "+Shift"
	}

	key += "+" + char

	return t.key(key, time, count, comment)
}

// Press the enter key.
func (t Tape) Enter(
	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Repeat the key press.
	//
	// +optional
	count int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.key(string(Enter), time, count, comment)
}

// Press the up arrow key.
func (t Tape) Up(
	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Repeat the key press.
	//
	// +optional
	count int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.key(string(Up), time, count, comment)
}

// Press the down arrow key.
func (t Tape) Down(
	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Repeat the key press.
	//
	// +optional
	count int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.key(string(Down), time, count, comment)
}

// Press the left arrow key.
func (t Tape) Left(
	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Repeat the key press.
	//
	// +optional
	count int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.key(string(Left), time, count, comment)
}

// Press the right arrow key.
func (t Tape) Right(
	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Repeat the key press.
	//
	// +optional
	count int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.key(string(Right), time, count, comment)
}

// Press the tab key.
func (t Tape) Tab(
	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Repeat the key press.
	//
	// +optional
	count int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.key(string(Tab), time, count, comment)
}

// Press the space key.
func (t Tape) Space(
	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Repeat the key press.
	//
	// +optional
	count int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.key(string(Space), time, count, comment)
}

// Press the page up key.
func (t Tape) PageUp(
	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Repeat the key press.
	//
	// +optional
	count int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.key(string(PageUp), time, count, comment)
}

// Press the page down key.
func (t Tape) PageDown(
	// Override the standard typing speed.
	//
	// +optional
	time string,

	// Repeat the key press.
	//
	// +optional
	count int,

	// Inline comment to add to the command.
	//
	// +optional
	comment string,
) Tape {
	return t.key(string(PageDown), time, count, comment)
}
