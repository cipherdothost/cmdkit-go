<!--
SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>

SPDX-License-Identifier: CC0-1.0
-->

# cmdkit

[![Go Documentation](https://pkg.go.dev/badge/v3.svg)](https://pkg.go.dev/go.cipher.host/cmdkit)
[![Go Report Card](https://goreportcard.com/badge/go.cipher.host/cmdkit)](https://goreportcard.com/report/go.cipher.host/cmdkit)
[![Tests status]( https://github.com/cipherdothost/cmdkit/actions/workflows/ci.yml/badge.svg)](https://github.com/cipherdothost/cmdkit/actions/workflows/ci.yml)

Package **cmdkit** provides a simple, intuitive, and kinda fun
command-line application framework that extends Go's `flag` package with
support for subcommands, better help formatting, and a simple-to-use
API.

The package features:

- Built on Go's standard `flag` package.
- No external dependencies.
- Subcommand support with proper flag handling.
- Unix-style short and long flags (`-h`, `--help`).
- Automatic help text generation and formatting.
- Before/After hooks for setup and cleanup.

The package also provide utilities that:

- Add environment variable support for flags.
- Add proper exit code handling.
- Add support for repeatable flags.
- Add smart terminal detection.
- Add color output detection with [NO_COLOR](https://no-color.org/)
  compliance.

## Installation

To install `cmdkit` and use it in your project, run:

```console
go get go.cipher.host/cmdkit@latest
```

## Usage

Here's a simple example that creates a greeting application:

```go
package main

import (
    "fmt"
    "os"

    "go.cipher.host/cmdkit"
)

func main() {
    // Create a new application
    app := cmdkit.New("greet", "A friendly greeting app", "1.0.0")

    // Add a "hello" command
    hello := cmdkit.NewCommand("hello", "Say hello to someone", "[--formal] <name>")
    
    var formal bool
    hello.Flags.BoolVar(&formal, "formal", false, "use formal greeting")
    
    hello.RunE = func(args []string) error {
        if len(args) < 1 {
            return fmt.Errorf("name is required")
        }

        if formal {
            fmt.Printf("Good day, %s!\n", args[0])
        } else {
            fmt.Printf("Hi, %s!\n", args[0])
        }

        return nil
    }

    app.AddCommand(hello)

    // Run the application
    cmdkit.Exit(app.Run(os.Args[1:]))
}
```

This will create a command-line application with the following interface:

```sh
$ greet --help
Usage: greet [global flags] command [command flags] [arguments...]

A friendly greeting app

Commands:
  hello         Say hello to someone

Global Flags:
  -h, --help    show help

Use "greet [command] --help" for more information about a command.

$ greet hello --help
Usage: greet hello [--formal] <name>

Say hello to someone

Flags:
  -h, --help     show help
  --formal       use formal greeting

$ greet hello Alice
Hi, Alice!

$ greet hello --formal Bob
Good day, Bob!
```

## Contributing

Anyone can help make **cmdkit** better. Check out [the contribution
guidelines](CONTRIBUTING.md) for more information.

---

Released under multiple licences and compliant with [the REUSE
specification](https://reuse.software/spec-3.3/). Please see the
individual files for details and [the LICENSES directory](LICENSES/) for
a full list of used licenses.
