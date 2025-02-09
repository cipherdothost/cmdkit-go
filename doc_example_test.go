// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit_test

import (
	"errors"
	"fmt"

	"go.cipher.host/cmdkit"
)

func Example() {
	var (
		// Create a new application with explicit metadata. You can also leave
		// the name and version empty to have the application infer them from
		// os.Args[0] and Go module information.
		app = cmdkit.New("myapp", "A simple example application", "1.0.0")

		// Add a simple command to the application.
		greet = cmdkit.NewCommand("greet", "Greet someone by name", "[--formal] <name>")
	)

	var formal bool

	// Register your flags using flag.TypeVar-kind of methods, e.g.
	// flag.BoolVar. If you want to add a Unix-style short form, register both
	// the long and short forms and then use App.AddShorthand.
	greet.Flags.BoolVar(&formal, "formal", false, "use formal greeting")

	// Set the command's run function.
	greet.RunE = func(args []string) error {
		if len(args) < 1 {
			return errors.New("name is required")
		}

		if formal {
			fmt.Fprintf(greet.OutWriter(), "Good day, %s!\n", args[0])
		} else {
			fmt.Fprintf(greet.OutWriter(), "Hi, %s!\n", args[0])
		}

		return nil
	}

	// Add the command to the application.
	app.AddCommand(greet)

	// Run the application with example arguments. You'd normally use
	// os.Args[1:] instead.
	if err := app.Run([]string{"greet", "Alice"}); err != nil {
		app.Errorf("error: %v", err)

		cmdkit.Exit(err)
	}

	// Output: Hi, Alice!
}
