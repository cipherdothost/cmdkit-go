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
	"strings"

	"go.cipher.host/cmdkit/internal/term"
)

// IsTerminal detects whether a given file descriptor represents an interactive
// terminal.
func IsTerminal(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}

// WantColor determines whether the application should use ANSI color codes in
// its output. It implements a comprehensive color detection system following
// modern terminal conventions:
//
// The function respects:
// - [NO_COLOR environment variable].
// - Terminal capability detection.
// - Special terminal types (e.g., "dumb" terminals).
//
// This provides a consistent and user-configurable color experience across
// different terminal environments.
//
// [NO_COLOR environment variable]: https://no-color.org/
func WantColor() bool {
	_, found := os.LookupEnv("NO_COLOR")

	return !found && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("TERM") != "dumb"
}

// LoadEnv bridges environment variables to command-line flags. It follows the
// common Unix pattern of using uppercase environment variables with a prefix.
//
// The function only sets flags that haven't been explicitly set via
// command-line arguments, maintaining the precedence:
// 1. Command-line flags.
// 2. Environment variables.
// 3. Default values.
//
// For example, for an app "gravity" with flag "--character",
// "GRAVITY_CHARACTER=mabel" would set the flag if not provided on command line.
func LoadEnv(fs *flag.FlagSet, prefix string) error {
	prefix = strings.ToUpper(prefix) + "_"

	var errs []error

	fs.VisitAll(func(f *flag.Flag) {
		if f.Value.String() != f.DefValue {
			return
		}

		envName := prefix + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		if envVal, ok := os.LookupEnv(envName); ok {
			if err := f.Value.Set(envVal); err != nil {
				errs = append(errs, fmt.Errorf("setting %q: %w", envName, err))
			}
		}
	})

	return errors.Join(errs...)
}

// InputOrFile returns a reader connected to stdin if path is "" or "-", or open
// a path for any other value. The Close method for stdin is a no-op.
//
// If message is given, it prints the message to stderr potentially notifying
// the user it's reading from stdin if the terminal is interactive.

// InputOrFile provides a flexible input mechanism that handles both
// file and stdin input transparently. It follows the Unix convention
// where "-" represents stdin, allowing commands to be part of pipelines.
//
// The function adds user-friendly behavior by displaying an optional
// message when reading from an interactive terminal, helping users
// understand that the program is waiting for input.
func InputOrFile(path, message string) (io.ReadCloser, error) {
	if path != "" && path != "-" {
		file, err := os.Open(path)
		if err != nil {
			err = fmt.Errorf("%w", err)
		}

		return file, err
	}

	if message != "" && IsTerminal(os.Stdin.Fd()) {
		fmt.Fprintf(os.Stderr, "%s\r", message)

		os.Stderr.Sync()
	}

	return io.NopCloser(os.Stdin), nil
}
