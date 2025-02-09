// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"

	"go.cipher.host/cmdkit/internal/xflag"
)

// ErrUnknownCommand is returned when a command is not found.
const ErrUnknownCommand Error = "unknown command"

// App represents a command-line application. It serves as the top-level
// container for application state and coordinates the execution of subcommands
// while managing global flags and configuration.
type App struct {
	// OutWriter receives normal output. When initialized from New, defaults to
	// [os.Stdout].
	//
	// Configurable primarily for testing and special output handling.
	//
	// [os.Stdout]: https://pkg.go.dev/os#Stdout
	OutWriter io.Writer

	// ErrWriter receives error output and help text. When initialized from New,
	// defaults to [os.Stderr].
	//
	// Separated from OutWriter to follow Unix output conventions.
	//
	// [os.Stderr]: https://pkg.go.dev/os#Stderr
	ErrWriter io.Writer

	// Flags holds the global flag set. These flags apply to all subcommands and
	// are parsed before command-specific flags.
	Flags *flag.FlagSet

	// Before is called before any subcommand execution, allowing for global
	// setup like configuration loading or resource initialization.
	Before func() error

	// After is called after subcommand execution, regardless of success or
	// failure. Useful for cleanup and resource release.
	After func() error

	// Name is the application name. When initialized from New, if not provided,
	// defaults to [os.Args[0]].
	//
	// [os.Args[0]]: https://pkg.go.dev/os#Args
	Name string

	// Description explains the application's purpose. Shown in help text to
	// guide users.
	Description string

	// Version indicates the application version. When initialized from new, if
	// not provided, defaults to [debug.ReadBuildInfo].
	//
	// [debug.ReadBuildInfo]: https://pkg.go.dev/runtime/debug#ReadBuildInfo
	Version string

	// Examples provides usage examples shown in help text. Each example should
	// be a complete command line.
	Examples []string

	// Commands holds the registered subcommands.
	Commands []*Command

	// shorthands maps short flag names, e.g. "-h", to their long versions, e.g.
	// "--help". This enables Unix-style short flags while maintaining clean
	// flag names.
	shorthands []xflag.Shorthand

	// helpFlag indicates whether to show helpFlag text.
	helpFlag bool

	// versionFlag indicates whether to show version information.
	versionFlag bool
}

// New creates a new App instance with sane defaults.
//
// Name resolution:
//   - Uses provided name if non-empty.
//   - Falls back to base name of [os.Args[0]] if name is empty.
//
// Version resolution:
//   - Uses provided version if non-empty.
//   - Attempts to read version from Go module information.
//   - Falls back to "unknown" if no version information is available.
//
// [os.Args[0]]: https://pkg.go.dev/os#Args
func New(name, description, version string) *App {
	if name == "" {
		name = filepath.Base(os.Args[0])
	}

	if version == "" {
		build, ok := debug.ReadBuildInfo()
		if ok {
			version = build.Main.Version
		} else {
			version = "unknown"
		}
	}

	app := &App{
		Name:        filepath.Base(name),
		Description: description,
		Version:     version,
		OutWriter:   os.Stdout,
		ErrWriter:   os.Stderr,
		Flags:       flag.NewFlagSet(name, flag.ContinueOnError),
		helpFlag:    false,
	}

	app.Flags.BoolVar(&app.helpFlag, "help", false, "show help")
	app.Flags.BoolVar(&app.helpFlag, "h", false, "show help")
	app.AddShorthand("help", "h")

	app.Flags.BoolVar(&app.versionFlag, "version", false, "show version")
	app.Flags.BoolVar(&app.versionFlag, "v", false, "show version")
	app.AddShorthand("version", "v")

	app.Flags.Usage = app.ShowHelp

	return app
}

// AddCommand registers a new subcommand with the application. Commands are
// stored in registration order, which affects help text display.
func (a *App) AddCommand(cmd *Command) {
	cmd.app = a

	a.Commands = append(a.Commands, cmd)
}

// AddShorthand creates a Unix-style short form for a flag. This enables the
// common pattern where flags have both long and short forms (e.g., "--help" and
// "-h").
//
// You should register both the long and short forms with flag.TypeVar-kind of
// methods, e.g. [flag.StringVar] before using this method.
//
// [flag.StringVar]: https://pkg.go.dev/flag#FlagSet.StringVar
func (a *App) AddShorthand(name, shorthand string) {
	if f := a.Flags.Lookup(name); f != nil {
		a.shorthands = append(a.shorthands, xflag.Shorthand{
			Name:      name,
			Shorthand: shorthand,
			Value:     f.Value,
		})
	}
}

// Run executes the application with the provided arguments, typically
// [os.Args[1:]].
//
// [os.Args[1:]]: https://pkg.go.dev/os#Args
func (a *App) Run(args []string) error {
	a.Flags.SetOutput(a.ErrWriter)

	if err := a.Flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return fmt.Errorf("%w", err)
	}

	if a.helpFlag {
		a.ShowHelp()

		return nil
	}

	if a.versionFlag {
		a.ShowVersion()

		return nil
	}

	if a.Before != nil {
		if err := a.Before(); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	defer func() {
		if a.After != nil {
			_ = a.After() //nolint:errcheck // not much we can do here
		}
	}()

	if a.Flags.NArg() == 0 {
		a.ShowHelp()

		return nil
	}

	cmdName := a.Flags.Arg(0)

	for _, cmd := range a.Commands {
		if cmd.Name == cmdName {
			return cmd.Run(a.Flags.Args()[1:])
		}
	}

	return fmt.Errorf("%w: %q", ErrUnknownCommand, cmdName)
}

// Errorf writes a formatted error message to the application's error output. It
// prefixes the message with the application name to provide context in error
// messages.
func (a *App) Errorf(format string, args ...any) {
	fmt.Fprintf(a.ErrWriter, "%s: %s\n", a.Name, fmt.Sprintf(format, args...))
}

// ShowVersion displays the application's version.
func (a *App) ShowVersion() {
	fmt.Fprintf(a.ErrWriter, "%s %s\n", a.Name, a.Version)
}

// ShowHelp displays the application's help text.
func (a *App) ShowHelp() {
	fmt.Fprintf(a.ErrWriter, "Usage: %s [global flags] command [command flags] [arguments...]\n\n", a.Name)

	if a.Description != "" {
		fmt.Fprintf(a.ErrWriter, "%s\n\n", a.Description)
	}

	if len(a.Commands) > 0 {
		fmt.Fprintf(a.ErrWriter, "Commands:\n")

		var (
			names  = make([]string, 0, len(a.Commands))
			cmdMap = make(map[string]*Command, len(a.Commands))
		)

		for _, cmd := range a.Commands {
			names = append(names, cmd.Name)

			cmdMap[cmd.Name] = cmd
		}

		sort.Strings(names)

		for _, name := range names {
			cmd := cmdMap[name]

			fmt.Fprintf(a.ErrWriter, "  %-12s %s\n", name, cmd.Description)
		}

		fmt.Fprintln(a.ErrWriter)
	}

	fmt.Fprintf(a.ErrWriter, "Global Flags:\n")
	a.Flags.SetOutput(a.ErrWriter)
	xflag.PrintFlags(a.ErrWriter, a.Flags, a.shorthands)

	fmt.Fprintln(a.ErrWriter)

	if len(a.Examples) > 0 {
		fmt.Fprintf(a.ErrWriter, "Examples:\n")

		for _, example := range a.Examples {
			fmt.Fprintf(a.ErrWriter, "  $ %s\n", example)
		}

		fmt.Fprintln(a.ErrWriter)
	}

	fmt.Fprintf(a.ErrWriter, "Use \"%s [command] --help\" for more information about a command.\n", a.Name)
}
