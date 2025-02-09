// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit_test

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"runtime/debug"
	"testing"

	"go.cipher.host/cmdkit"
)

func TestNew(t *testing.T) { //nolint:paralleltest // not thread-safe
	originalArgs := os.Args

	t.Cleanup(func() {
		os.Args = originalArgs
	})

	var (
		buildInfo, _   = debug.ReadBuildInfo()
		defaultVersion = "unknown"
	)

	if buildInfo != nil {
		defaultVersion = buildInfo.Main.Version
	}

	tests := []struct {
		name           string
		giveName       string
		giveDesc       string
		giveVersion    string
		setupArgs      []string
		wantName       string
		wantDesc       string
		wantVersion    string
		wantFlagNames  []string
		wantShorthands map[string]string
	}{
		{
			name:        "all values provided",
			giveName:    "testapp",
			giveDesc:    "Test Description",
			giveVersion: "1.0.0",
			wantName:    "testapp",
			wantDesc:    "Test Description",
			wantVersion: "1.0.0",
			wantFlagNames: []string{
				"help",
				"h",
				"version",
				"v",
			},
			wantShorthands: map[string]string{
				"help":    "h",
				"version": "v",
			},
		},
		{
			name:        "empty name uses binary name",
			giveName:    "",
			giveDesc:    "Test Description",
			giveVersion: "1.0.0",
			setupArgs: []string{
				"/usr/local/bin/myapp",
			},
			wantName:    "myapp",
			wantDesc:    "Test Description",
			wantVersion: "1.0.0",
			wantFlagNames: []string{
				"help",
				"h",
				"version",
				"v",
			},
			wantShorthands: map[string]string{
				"help":    "h",
				"version": "v",
			},
		},
		{
			name:        "empty version uses build info",
			giveName:    "testapp",
			giveDesc:    "Test Description",
			giveVersion: "",
			wantName:    "testapp",
			wantDesc:    "Test Description",
			wantVersion: defaultVersion,
			wantFlagNames: []string{
				"help",
				"h",
				"version",
				"v",
			},
			wantShorthands: map[string]string{
				"help":    "h",
				"version": "v",
			},
		},
		{
			name:        "minimal configuration",
			giveName:    "",
			giveDesc:    "",
			giveVersion: "",
			setupArgs: []string{
				"/usr/local/bin/myapp",
			},
			wantName:    "myapp",
			wantDesc:    "",
			wantVersion: defaultVersion,
			wantFlagNames: []string{
				"help",
				"h",
				"version",
				"v",
			},
			wantShorthands: map[string]string{
				"help":    "h",
				"version": "v",
			},
		},
	}

	for _, tt := range tests { //nolint:paralleltest // not thread-safe
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupArgs != nil {
				os.Args = tt.setupArgs
			}

			app := cmdkit.New(tt.giveName, tt.giveDesc, tt.giveVersion)

			if got := app.Name; got != tt.wantName {
				t.Errorf("New() Name = %q, want %q", got, tt.wantName)
			}

			if got := app.Description; got != tt.wantDesc {
				t.Errorf("New() Description = %q, want %q", got, tt.wantDesc)
			}

			if got := app.Version; got != tt.wantVersion {
				t.Errorf("New() Version = %q, want %q", got, tt.wantVersion)
			}

			if app.OutWriter == nil {
				t.Error("New() OutWriter is nil")
			}

			if app.ErrWriter == nil {
				t.Error("New() ErrWriter is nil")
			}

			if app.Flags == nil {
				t.Fatal("New() Flags is nil")
			}

			seenFlags := make(map[string]bool)

			app.Flags.VisitAll(func(f *flag.Flag) {
				seenFlags[f.Name] = true
			})

			for _, flagName := range tt.wantFlagNames {
				if !seenFlags[flagName] {
					t.Errorf("New() missing expected flag %q", flagName)
				}
			}

			var buf bytes.Buffer
			app.ErrWriter = &buf

			app.ShowVersion()

			wantVersion := tt.wantName + " " + tt.wantVersion + "\n"
			if got := buf.String(); got != wantVersion {
				t.Errorf("New() version output = %q, want %q", got, wantVersion)
			}

			buf.Reset()

			app.ShowHelp()

			if !bytes.Contains(buf.Bytes(), []byte("Usage:")) {
				t.Error("New() help output missing Usage section")
			}

			for longName, shortName := range tt.wantShorthands {
				var (
					longFlag  = app.Flags.Lookup(longName)
					shortFlag = app.Flags.Lookup(shortName)
				)

				if longFlag == nil {
					t.Errorf("New() long flag %q not found", longName)

					continue
				}

				if shortFlag == nil {
					t.Errorf("New() short flag %q not found", shortName)

					continue
				}

				if err := longFlag.Value.Set("true"); err != nil {
					t.Errorf("New() setting %q: %v", longName, err)
				}

				if shortFlag.Value.String() != "true" {
					t.Errorf("New() shorthand %q not properly linked to %q", shortName, longName)
				}
			}
		})
	}
}

func TestApp_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupApp      func(*testing.T) *cmdkit.App
		giveArgs      []string
		wantErr       error
		wantOutput    string
		wantErrOutput string
		wantHookOrder []string
	}{
		{
			name: "basic command execution",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				var (
					app = cmdkit.New("testapp", "Test App", "1.0.0")
					cmd = cmdkit.NewCommand("greet", "Greeting command", "")
				)

				cmd.RunE = func(_ []string) error {
					return nil
				}

				app.AddCommand(cmd)

				return app
			},
			giveArgs: []string{
				"greet",
			},
		},
		{
			name: "help flag shows help",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				return cmdkit.New("testapp", "Test App", "1.0.0")
			},
			giveArgs: []string{
				"--help",
			},
			wantErrOutput: "Usage: testapp [global flags] command [command flags] [arguments...]\n\n" +
				"Test App\n\n" +
				"Global Flags:\n" +
				"  -h, --help     show help (defaults to \"false\")\n" +
				"  -v, --version  show version (defaults to \"false\")\n\n" +
				"Use \"testapp [command] --help\" for more information about a command.\n",
		},
		{
			name: "version flag shows version",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				return cmdkit.New("testapp", "Test App", "1.0.0")
			},
			giveArgs: []string{
				"--version",
			},
			wantErrOutput: "testapp 1.0.0\n",
		},
		{
			name: "unknown command",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				return cmdkit.New("testapp", "Test App", "1.0.0")
			},
			giveArgs: []string{
				"unknown",
			},
			wantErr: errors.New(`unknown command: "unknown"`),
		},
		{
			name: "no command shows help",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				return cmdkit.New("testapp", "Test App", "1.0.0")
			},
			giveArgs: make([]string, 0),
			wantErrOutput: "Usage: testapp [global flags] command [command flags] [arguments...]\n\n" +
				"Test App\n\n" +
				"Global Flags:\n" +
				"  -h, --help     show help (defaults to \"false\")\n" +
				"  -v, --version  show version (defaults to \"false\")\n\n" +
				"Use \"testapp [command] --help\" for more information about a command.\n",
		},
		{
			name: "before hook error",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				app := cmdkit.New("testapp", "Test App", "1.0.0")
				app.Before = func() error {
					return errors.New("before hook error")
				}

				return app
			},
			giveArgs: []string{
				"command",
			},
			wantErr: errors.New("before hook error"),
		},
		{
			name: "hook execution order",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				var order []string

				app := cmdkit.New("testapp", "Test App", "1.0.0")
				app.Before = func() error {
					order = append(order, "before")

					return nil
				}
				app.After = func() error {
					order = append(order, "after")

					return nil
				}

				cmd := cmdkit.NewCommand("test", "Test command", "")
				cmd.RunE = func(_ []string) error {
					order = append(order, "command")

					return nil
				}

				app.AddCommand(cmd)

				t.Cleanup(func() {
					want := []string{"before", "command", "after"}

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

				return app
			},
			giveArgs: []string{
				"test",
			},
		},
		{
			name: "command error propagation",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				var (
					app = cmdkit.New("testapp", "Test App", "1.0.0")
					cmd = cmdkit.NewCommand("fail", "Failing command", "")
				)

				cmd.RunE = func(_ []string) error {
					return errors.New("command error")
				}

				app.AddCommand(cmd)

				return app
			},
			giveArgs: []string{
				"fail",
			},
			wantErr: errors.New("command error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := tt.setupApp(t)

			var stdout, stderr bytes.Buffer
			app.OutWriter = &stdout
			app.ErrWriter = &stderr

			err := app.Run(tt.giveArgs)
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

func TestApp_Errorf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupApp   func(*testing.T) *cmdkit.App
		giveFormat string
		giveArgs   []any
		want       string
	}{
		{
			name: "simple message",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				return cmdkit.New("testapp", "", "1.0.0")
			},
			giveFormat: "something went wrong",
			giveArgs:   nil,
			want:       "testapp: something went wrong\n",
		},
		{
			name: "formatted message",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				return cmdkit.New("testapp", "", "1.0.0")
			},
			giveFormat: "error: %v",
			giveArgs: []any{
				errors.New("failed to do something"),
			},
			want: "testapp: error: failed to do something\n",
		},
		{
			name: "multiple arguments",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				return cmdkit.New("testapp", "", "1.0.0")
			},
			giveFormat: "count: %d, name: %q",
			giveArgs: []any{
				42,
				"test",
			},
			want: "testapp: count: 42, name: \"test\"\n",
		},
		{
			name: "custom app name",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				return cmdkit.New("myapp", "", "1.0.0")
			},
			giveFormat: "something failed",
			giveArgs:   nil,
			want:       "myapp: something failed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := tt.setupApp(t)

			var stderr bytes.Buffer
			app.ErrWriter = &stderr

			app.Errorf(tt.giveFormat, tt.giveArgs...)

			if got := stderr.String(); got != tt.want {
				t.Errorf("Errorf() output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApp_ShowHelp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setupApp func(*testing.T) *cmdkit.App
		want     string
	}{
		{
			name: "minimal app",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				return cmdkit.New("testapp", "", "1.0.0")
			},
			want: "Usage: testapp [global flags] command [command flags] [arguments...]\n\n" +
				"Global Flags:\n" +
				"  -h, --help     show help (defaults to \"false\")\n" +
				"  -v, --version  show version (defaults to \"false\")\n\n" +
				"Use \"testapp [command] --help\" for more information about a command.\n",
		},
		{
			name: "app with description",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				return cmdkit.New("testapp", "A test application", "1.0.0")
			},
			want: "Usage: testapp [global flags] command [command flags] [arguments...]\n\n" +
				"A test application\n\n" +
				"Global Flags:\n" +
				"  -h, --help     show help (defaults to \"false\")\n" +
				"  -v, --version  show version (defaults to \"false\")\n\n" +
				"Use \"testapp [command] --help\" for more information about a command.\n",
		},
		{
			name: "app with commands",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				app := cmdkit.New("testapp", "", "1.0.0")

				app.AddCommand(cmdkit.NewCommand("zebra", "Last command", ""))
				app.AddCommand(cmdkit.NewCommand("alpha", "First command", ""))
				app.AddCommand(cmdkit.NewCommand("beta", "Second command", ""))

				return app
			},
			want: "Usage: testapp [global flags] command [command flags] [arguments...]\n\n" +
				"Commands:\n" +
				"  alpha        First command\n" +
				"  beta         Second command\n" +
				"  zebra        Last command\n\n" +
				"Global Flags:\n" +
				"  -h, --help     show help (defaults to \"false\")\n" +
				"  -v, --version  show version (defaults to \"false\")\n\n" +
				"Use \"testapp [command] --help\" for more information about a command.\n",
		},
		{
			name: "app with examples",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				app := cmdkit.New("testapp", "", "1.0.0")
				app.Examples = []string{
					"testapp greet John",
					"testapp greet --formal Mary",
				}

				return app
			},
			want: "Usage: testapp [global flags] command [command flags] [arguments...]\n\n" +
				"Global Flags:\n" +
				"  -h, --help     show help (defaults to \"false\")\n" +
				"  -v, --version  show version (defaults to \"false\")\n\n" +
				"Examples:\n" +
				"  $ testapp greet John\n" +
				"  $ testapp greet --formal Mary\n\n" +
				"Use \"testapp [command] --help\" for more information about a command.\n",
		},
		{
			name: "full featured app",
			setupApp: func(t *testing.T) *cmdkit.App {
				t.Helper()

				app := cmdkit.New("testapp", "A fully featured test application", "1.0.0")

				app.AddCommand(cmdkit.NewCommand("greet", "Greet a person", ""))
				app.AddCommand(cmdkit.NewCommand("config", "Configure the application", ""))

				app.Examples = []string{
					"testapp greet John",
					"testapp config --set key=value",
				}

				return app
			},
			want: "Usage: testapp [global flags] command [command flags] [arguments...]\n\n" +
				"A fully featured test application\n\n" +
				"Commands:\n" +
				"  config       Configure the application\n" +
				"  greet        Greet a person\n\n" +
				"Global Flags:\n" +
				"  -h, --help     show help (defaults to \"false\")\n" +
				"  -v, --version  show version (defaults to \"false\")\n\n" +
				"Examples:\n" +
				"  $ testapp greet John\n" +
				"  $ testapp config --set key=value\n\n" +
				"Use \"testapp [command] --help\" for more information about a command.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := tt.setupApp(t)

			var stderr bytes.Buffer
			app.ErrWriter = &stderr

			app.ShowHelp()

			if got := stderr.String(); got != tt.want {
				t.Errorf("ShowHelp() output = %q, want %q", got, tt.want)
			}
		})
	}
}
