// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit_test

import (
	"bytes"
	"errors"
	"flag"
	"testing"

	"go.cipher.host/cmdkit"
)

func TestNewCommand(t *testing.T) { //nolint:paralleltest // not thread-safe
	tests := []struct {
		name           string
		giveName       string
		giveDesc       string
		giveUsage      string
		wantName       string
		wantDesc       string
		wantUsage      string
		wantFlagNames  []string
		wantShorthands map[string]string
		wantHelpText   string
	}{
		{
			name:      "all values provided",
			giveName:  "test",
			giveDesc:  "Test Description",
			giveUsage: "[flags] <input>",
			wantName:  "test",
			wantDesc:  "Test Description",
			wantUsage: "[flags] <input>",
			wantFlagNames: []string{
				"help",
				"h",
			},
			wantShorthands: map[string]string{
				"help": "h",
			},
			wantHelpText: "Usage: testapp test [flags] <input>\n\n" +
				"Test Description\n\n" +
				"Flags:\n" +
				"  -h, --help  show help (defaults to \"false\")\n\n",
		},
		{
			name:      "minimal configuration",
			giveName:  "minimal",
			giveDesc:  "",
			giveUsage: "",
			wantName:  "minimal",
			wantDesc:  "",
			wantUsage: "",
			wantFlagNames: []string{
				"help",
				"h",
			},
			wantShorthands: map[string]string{
				"help": "h",
			},
			wantHelpText: "Usage: testapp minimal\n\n" +
				"Flags:\n" +
				"  -h, --help  show help (defaults to \"false\")\n\n",
		},
		{
			name:      "with examples",
			giveName:  "greet",
			giveDesc:  "Greet someone",
			giveUsage: "[flags] <name>",
			wantName:  "greet",
			wantDesc:  "Greet someone",
			wantUsage: "[flags] <name>",
			wantFlagNames: []string{
				"help",
				"h",
			},
			wantShorthands: map[string]string{
				"help": "h",
			},
			wantHelpText: "Usage: testapp greet [flags] <name>\n\n" +
				"Greet someone\n\n" +
				"Flags:\n" +
				"  -h, --help  show help (defaults to \"false\")\n\n",
		},
	}

	for _, tt := range tests { //nolint:paralleltest // not thread-safe
		t.Run(tt.name, func(t *testing.T) {
			cmd := cmdkit.NewCommand(tt.giveName, tt.giveDesc, tt.giveUsage)

			if got := cmd.Name; got != tt.wantName {
				t.Errorf("NewCommand() Name = %q, want %q", got, tt.wantName)
			}

			if got := cmd.Description; got != tt.wantDesc {
				t.Errorf("NewCommand() Description = %q, want %q", got, tt.wantDesc)
			}

			if got := cmd.Usage; got != tt.wantUsage {
				t.Errorf("NewCommand() Usage = %q, want %q", got, tt.wantUsage)
			}

			if cmd.Flags == nil {
				t.Fatal("NewCommand() Flags is nil")
			}

			seenFlags := make(map[string]bool)

			cmd.Flags.VisitAll(func(f *flag.Flag) {
				seenFlags[f.Name] = true
			})

			for _, flagName := range tt.wantFlagNames {
				if !seenFlags[flagName] {
					t.Errorf("NewCommand() missing expected flag %q", flagName)
				}
			}

			app := cmdkit.New("testapp", "", "1.0.0")
			app.AddCommand(cmd)

			var buf bytes.Buffer
			app.ErrWriter = &buf

			cmd.ShowHelp()

			if got := buf.String(); got != tt.wantHelpText {
				t.Errorf("NewCommand() help text = %q, want %q", got, tt.wantHelpText)
			}

			for longName, shortName := range tt.wantShorthands {
				var (
					longFlag  = cmd.Flags.Lookup(longName)
					shortFlag = cmd.Flags.Lookup(shortName)
				)

				if longFlag == nil {
					t.Errorf("NewCommand() long flag %q not found", longName)

					continue
				}

				if shortFlag == nil {
					t.Errorf("NewCommand() short flag %q not found", shortName)

					continue
				}

				if err := longFlag.Value.Set("true"); err != nil {
					t.Errorf("NewCommand() setting %q: %v", longName, err)
				}

				if shortFlag.Value.String() != "true" {
					t.Errorf("NewCommand() shorthand %q not properly linked to %q", shortName, longName)
				}
			}
		})
	}
}

func TestCommand_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupCmd      func(*testing.T) *cmdkit.Command
		giveArgs      []string
		wantErr       error
		wantOutput    string
		wantErrOutput string
		wantHookOrder []string
	}{
		{
			name: "basic command execution",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				cmd := cmdkit.NewCommand("test", "Test command", "")
				cmd.RunE = func(_ []string) error {
					return nil
				}

				return cmd
			},
			giveArgs: make([]string, 0),
		},
		{
			name: "help flag shows help",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				return cmdkit.NewCommand("test", "Test command", "[flags] <input>")
			},
			giveArgs: []string{
				"--help",
			},
			wantErrOutput: "Usage: testapp test [flags] <input>\n\n" +
				"Test command\n\n" +
				"Flags:\n" +
				"  -h, --help  show help (defaults to \"false\")\n\n",
		},
		{
			name: "help shorthand flag shows help",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				return cmdkit.NewCommand("test", "Test command", "[flags] <input>")
			},
			giveArgs: []string{
				"-h",
			},
			wantErrOutput: "Usage: testapp test [flags] <input>\n\n" +
				"Test command\n\n" +
				"Flags:\n" +
				"  -h, --help  show help (defaults to \"false\")\n\n",
		},
		{
			name: "before hook error",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				cmd := cmdkit.NewCommand("test", "Test command", "")
				cmd.Before = func() error {
					return errors.New("before hook error")
				}

				return cmd
			},
			giveArgs: make([]string, 0),
			wantErr:  errors.New("before hook error"),
		},
		{
			name: "after hook error is ignored",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				cmd := cmdkit.NewCommand("test", "Test command", "")
				cmd.RunE = func(_ []string) error {
					return nil
				}
				cmd.After = func() error {
					return errors.New("after hook error")
				}

				return cmd
			},
			giveArgs: make([]string, 0),
		},
		{
			name: "hook execution order",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				var order []string

				cmd := cmdkit.NewCommand("test", "Test command", "")
				cmd.Before = func() error {
					order = append(order, "before")

					return nil
				}
				cmd.RunE = func(_ []string) error {
					order = append(order, "run")
					return nil
				}
				cmd.After = func() error {
					order = append(order, "after")

					return nil
				}

				t.Cleanup(func() {
					want := []string{"before", "run", "after"}

					if len(order) != len(want) {
						t.Errorf("Run() hook order length = %d, want %d", len(order), len(want))

						return
					}

					for i := range order {
						if order[i] != want[i] {
							t.Errorf("Run() hook order[%d] = %q, want %q", i, order[i], want[i])
						}
					}
				})

				return cmd
			},
			giveArgs: make([]string, 0),
		},
		{
			name: "missing RunE function",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				return cmdkit.NewCommand("test", "Test command", "")
			},
			giveArgs: make([]string, 0),
			wantErr:  errors.New(`"test" command has no run function`),
		},
		{
			name: "command with flags",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				cmd := cmdkit.NewCommand("test", "Test command", "")

				var verbose bool

				cmd.Flags.BoolVar(&verbose, "verbose", false, "verbose output")
				cmd.Flags.BoolVar(&verbose, "v", false, "verbose output")

				cmd.AddShorthand("verbose", "v")

				cmd.RunE = func(_ []string) error {
					if verbose {
						return errors.New("verbose mode")
					}

					return nil
				}

				return cmd
			},
			giveArgs: []string{
				"--verbose",
			},
			wantErr: errors.New("verbose mode"),
		},
		{
			name: "command with arguments",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				cmd := cmdkit.NewCommand("test", "Test command", "")
				cmd.RunE = func(args []string) error {
					if len(args) != 1 || args[0] != "arg1" {
						return errors.New("unexpected arguments")
					}

					return nil
				}

				return cmd
			},
			giveArgs: []string{
				"arg1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := tt.setupCmd(t)

			app := cmdkit.New("testapp", "", "1.0.0")
			app.AddCommand(cmd)

			var stdout, stderr bytes.Buffer
			app.OutWriter = &stdout
			app.ErrWriter = &stderr

			if cmd.OutWriter() != app.OutWriter {
				t.Errorf("Run() OutWriter = %v, want %v", cmd.OutWriter(), app.OutWriter)
			}

			if cmd.ErrWriter() != app.ErrWriter {
				t.Errorf("Run() ErrWriter = %v, want %v", cmd.ErrWriter(), app.ErrWriter)
			}

			err := cmd.Run(tt.giveArgs)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Run() error = nil, want error %v", tt.wantErr)
				}

				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Errorf("Run() error = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Run() unexpected error = %v", err)
			}

			if got := stdout.String(); got != tt.wantOutput {
				t.Errorf("Run() stdout = %q, want %q", got, tt.wantOutput)
			}

			if got := stderr.String(); got != tt.wantErrOutput {
				t.Errorf("Run() stderr = %q, want %q", got, tt.wantErrOutput)
			}
		})
	}
}

func TestCommand_ShowHelp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setupCmd func(*testing.T) *cmdkit.Command
		want     string
	}{
		{
			name: "minimal command",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				return cmdkit.NewCommand("test", "", "")
			},
			want: "Usage: testapp test\n\n" +
				"Flags:\n" +
				"  -h, --help  show help (defaults to \"false\")\n\n",
		},
		{
			name: "command with description",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				return cmdkit.NewCommand("test", "A test command", "")
			},
			want: "Usage: testapp test\n\n" +
				"A test command\n\n" +
				"Flags:\n" +
				"  -h, --help  show help (defaults to \"false\")\n\n",
		},
		{
			name: "command with usage",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				return cmdkit.NewCommand("test", "", "[flags] <input> <output>")
			},
			want: "Usage: testapp test [flags] <input> <output>\n\n" +
				"Flags:\n" +
				"  -h, --help  show help (defaults to \"false\")\n\n",
		},
		{
			name: "command with custom flags",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				cmd := cmdkit.NewCommand("test", "", "")

				var (
					verbose bool
					count   int
					name    string
				)

				cmd.Flags.BoolVar(&verbose, "verbose", false, "enable verbose output")
				cmd.Flags.IntVar(&count, "count", 0, "number of items")
				cmd.Flags.StringVar(&name, "name", "", "item name")

				return cmd
			},
			want: "Usage: testapp test\n\n" +
				"Flags:\n" +
				"  -count      number of items (defaults to \"0\")\n" +
				"  -h, --help  show help (defaults to \"false\")\n" +
				"  -name       item name\n" +
				"  -verbose    enable verbose output (defaults to \"false\")\n\n",
		},
		{
			name: "command with examples",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				cmd := cmdkit.NewCommand("test", "", "")
				cmd.Examples = []string{
					"testapp test input.txt",
					"testapp test --verbose input.txt output.txt",
				}

				return cmd
			},
			want: "Usage: testapp test\n\n" +
				"Flags:\n" +
				"  -h, --help  show help (defaults to \"false\")\n\n" +
				"Examples:\n" +
				"  $ testapp test input.txt\n" +
				"  $ testapp test --verbose input.txt output.txt\n",
		},
		{
			name: "command with shorthands",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				cmd := cmdkit.NewCommand("test", "", "")

				var (
					verbose bool
					output  string
				)

				cmd.Flags.BoolVar(&verbose, "verbose", false, "enable verbose output")
				cmd.Flags.BoolVar(&verbose, "v", false, "enable verbose output")
				cmd.AddShorthand("verbose", "v")

				cmd.Flags.StringVar(&output, "output", "", "output file")
				cmd.Flags.StringVar(&output, "o", "", "output file")
				cmd.AddShorthand("output", "o")

				return cmd
			},
			want: "Usage: testapp test\n\n" +
				"Flags:\n" +
				"  -h, --help     show help (defaults to \"false\")\n" +
				"  -o, --output   output file\n" +
				"  -v, --verbose  enable verbose output (defaults to \"false\")\n\n",
		},
		{
			name: "full featured command",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				cmd := cmdkit.NewCommand("process", "Process input files", "[flags] <input...>")

				var (
					verbose  bool
					output   string
					format   string
					workers  int
					dryRun   bool
					logFile  string
					compress bool
				)

				cmd.Flags.BoolVar(&verbose, "verbose", false, "enable verbose output")
				cmd.Flags.BoolVar(&verbose, "v", false, "enable verbose output")
				cmd.AddShorthand("verbose", "v")

				cmd.Flags.StringVar(&output, "output", "", "output file")
				cmd.Flags.StringVar(&output, "o", "", "output file")
				cmd.AddShorthand("output", "o")

				cmd.Flags.StringVar(&format, "format", "json", "output format")
				cmd.Flags.IntVar(&workers, "workers", 1, "number of workers")
				cmd.Flags.BoolVar(&dryRun, "dry-run", false, "don't actually process files")
				cmd.Flags.StringVar(&logFile, "log-file", "", "log file path")
				cmd.Flags.BoolVar(&compress, "compress", false, "compress output")

				cmd.Examples = []string{
					"testapp process input.txt",
					"testapp process -o output.json input1.txt input2.txt",
					"testapp process --format=xml --workers=4 *.txt",
				}

				return cmd
			},
			want: "Usage: testapp process [flags] <input...>\n\n" +
				"Process input files\n\n" +
				"Flags:\n" +
				"  -compress      compress output (defaults to \"false\")\n" +
				"  -dry-run       don't actually process files (defaults to \"false\")\n" +
				"  -format        output format (defaults to \"json\")\n" +
				"  -h, --help     show help (defaults to \"false\")\n" +
				"  -log-file      log file path\n" +
				"  -o, --output   output file\n" +
				"  -v, --verbose  enable verbose output (defaults to \"false\")\n" +
				"  -workers       number of workers (defaults to \"1\")\n\n" +
				"Examples:\n" +
				"  $ testapp process input.txt\n" +
				"  $ testapp process -o output.json input1.txt input2.txt\n" +
				"  $ testapp process --format=xml --workers=4 *.txt\n",
		},
		{
			name: "command with long flag names",
			setupCmd: func(t *testing.T) *cmdkit.Command {
				t.Helper()

				cmd := cmdkit.NewCommand("test", "", "")

				var (
					veryLongFlagName1 string
					veryLongFlagName2 bool
				)

				cmd.Flags.StringVar(&veryLongFlagName1, "very-long-flag-name-1", "", "first very long flag")
				cmd.Flags.BoolVar(&veryLongFlagName2, "very-long-flag-name-2", false, "second very long flag")

				return cmd
			},
			want: "Usage: testapp test\n\n" +
				"Flags:\n" +
				"  -h, --help             show help (defaults to \"false\")\n" +
				"  -very-long-flag-name-1 first very long flag\n" +
				"  -very-long-flag-name-2 second very long flag (defaults to \"false\")\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := tt.setupCmd(t)

			app := cmdkit.New("testapp", "", "1.0.0")
			app.AddCommand(cmd)

			var buf bytes.Buffer
			app.ErrWriter = &buf

			cmd.ShowHelp()

			if got := buf.String(); got != tt.want {
				t.Errorf("ShowHelp() output = %q, want %q", got, tt.want)
			}
		})
	}
}
