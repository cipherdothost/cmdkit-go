// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"go.cipher.host/cmdkit/internal/xflag"
)

// ErrCommandRunMissing is returned when a command is created without the RunE
// field.
var ErrCommandRunMissing Error = "command has no run function"

// Command represents a single subcommand in a command-line application. It
// encapsulates all the behavior and flags specific to one function of the
// application, following the Unix principle of small, focused tools.
//
// Commands are designed to be self-contained units that can be composed into
// larger applications. Each command manages its own flags and handles its own
// argument validation and processing.
type Command struct {
	// Flags holds command-specific flags that only apply to this command. These
	// are parsed after global flags but before command arguments.
	Flags *flag.FlagSet

	// RunE executes the command's logic. It receives the remaining arguments
	// after flag parsing.
	RunE func(args []string) error

	// Before runs before command execution, allowing for command-specific setup
	// like validating flags or preparing resources.
	Before func() error

	// After runs after command execution for cleanup. Executes even if the
	// command fails.
	After func() error

	// app is the parent application, providing access to global state and
	// configuration.
	app *App

	// Name is the command name as used in the command line.
	Name string

	// Description explains the command's purpose. Shown in both parent app help
	// and command-specific help.
	Description string

	// Usage describes the command's argument pattern. E.g., "[flags] <input>
	// <output>".
	Usage string

	// Examples shows example invocations of this specific command.
	Examples []string

	// shorthands maps short flag names, e.g. "-h", to their long versions, e.g.
	// "--help". This enables Unix-style short flags while maintaining clean
	// flag names.
	shorthands []xflag.Shorthand

	// helpFlag indicates whether to show helpFlag text.
	helpFlag bool
}

// NewCommand creates a new Command instance with the provided metadata.
//
// For the best user experience, the usage string should follow Unix
// conventions:
// - Optional elements in square brackets: [--flag].
// - Required elements in angle brackets: <file>.
// - Variable numbers of arguments with ellipsis: [file...].
func NewCommand(name, description, usage string) *Command {
	cmd := &Command{
		Name:        name,
		Description: description,
		Usage:       usage,
		Flags:       flag.NewFlagSet(name, flag.ContinueOnError),
	}

	cmd.Flags.BoolVar(&cmd.helpFlag, "help", false, "show help")
	cmd.Flags.BoolVar(&cmd.helpFlag, "h", false, "show help")
	cmd.AddShorthand("help", "h")

	cmd.Flags.Usage = cmd.ShowHelp

	return cmd
}

// AddShorthand creates a short form alias for a command-specific flag. This
// works the same way as [App.AddShorthand] but scoped to the command's flags.
//
// [app.AddShorthand]: https://pkg.go.dev/go.cipher.host/cmdkit#App.AddShorthand
func (c *Command) AddShorthand(name, shorthand string) {
	if f := c.Flags.Lookup(name); f != nil {
		c.shorthands = append(c.shorthands, xflag.Shorthand{
			Name:      name,
			Shorthand: shorthand,
			Value:     f.Value,
		})
	}
}

// Run executes the command with the provided arguments. It follows
// a similar flow to [App.Run].
//
// [App.Run]: https://pkg.go.dev/go.cipher.host/cmdkit#App.Run
func (c *Command) Run(args []string) error {
	c.Flags.SetOutput(c.app.ErrWriter)

	if err := c.Flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return fmt.Errorf("%w", err)
	}

	if c.helpFlag {
		c.ShowHelp()

		return nil
	}

	if c.Before != nil {
		if err := c.Before(); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	defer func() {
		if c.After != nil {
			_ = c.After() //nolint:errcheck // not much we can do here
		}
	}()

	if c.RunE == nil {
		return fmt.Errorf("%q %w", c.Name, ErrCommandRunMissing)
	}

	return c.RunE(c.Flags.Args())
}

// OutWriter returns the application's output writer.
func (c *Command) OutWriter() io.Writer {
	return c.app.OutWriter
}

// ErrWriter returns the application's error writer.
func (c *Command) ErrWriter() io.Writer {
	return c.app.ErrWriter
}

// ShowHelp displays command-specific help text.
func (c *Command) ShowHelp() {
	fmt.Fprintf(c.app.ErrWriter, "Usage: %s %s", c.app.Name, c.Name)

	if c.Usage != "" {
		fmt.Fprintf(c.app.ErrWriter, " %s", c.Usage)
	}

	fmt.Fprintln(c.app.ErrWriter)
	fmt.Fprintln(c.app.ErrWriter)

	if c.Description != "" {
		fmt.Fprintf(c.app.ErrWriter, "%s\n\n", c.Description)
	}

	fmt.Fprintf(c.app.ErrWriter, "Flags:\n")
	c.Flags.SetOutput(c.app.ErrWriter)
	xflag.PrintFlags(c.app.ErrWriter, c.Flags, c.shorthands)

	fmt.Fprintln(c.app.ErrWriter)

	if len(c.Examples) > 0 {
		fmt.Fprintf(c.app.ErrWriter, "Examples:\n")

		for _, example := range c.Examples {
			fmt.Fprintf(c.app.ErrWriter, "  $ %s\n", example)
		}
	}
}
