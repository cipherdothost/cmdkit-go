// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit_test

import (
	"flag"
	"os"
	"runtime"
	"testing"

	"go.cipher.host/cmdkit"
)

func TestIsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		give func(t *testing.T) uintptr
		want bool
	}{
		{
			name: "regular file",
			give: func(t *testing.T) uintptr {
				t.Helper()

				file, err := os.CreateTemp(t.TempDir(), "TestIsTerminal")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}

				t.Cleanup(func() {
					file.Close()
				})

				return file.Fd()
			},
			want: false,
		},
		{
			name: "invalid file descriptor",
			give: func(t *testing.T) uintptr {
				t.Helper()

				return ^uintptr(0)
			},
			want: false,
		},
		{
			name: "stdin",
			give: func(t *testing.T) uintptr {
				t.Helper()

				return os.Stdin.Fd()
			},
			want: false,
		},
		{
			name: "stdout",
			give: func(t *testing.T) uintptr {
				t.Helper()

				return os.Stdout.Fd()
			},
			want: false,
		},
		{
			name: "stderr",
			give: func(t *testing.T) uintptr {
				t.Helper()

				return os.Stderr.Fd()
			},
			want: false,
		},
	}

	if runtime.GOOS == "linux" {
		tests = append(tests, struct {
			name string
			give func(t *testing.T) uintptr
			want bool
		}{
			name: "terminal device",
			give: func(t *testing.T) uintptr {
				t.Helper()

				file, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
				if err != nil {
					t.Fatalf("failed to open /dev/ptmx: %v", err)
				}

				t.Cleanup(func() {
					file.Close()
				})

				return file.Fd()
			},
			want: true,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				fd  = tt.give(t)
				got = cmdkit.IsTerminal(fd)
			)

			if got != tt.want {
				t.Errorf("IsTerminal(%d) = %v, want %v", fd, got, tt.want)
			}
		})
	}
}

func TestWantColor(t *testing.T) { //nolint:paralleltest // t.Setenv is not thread-safe
	tests := []struct {
		name  string
		setup func(t *testing.T)
		want  bool
	}{
		{
			name: "NO_COLOR set",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("NO_COLOR", "1")
			},
			want: false,
		},
		{
			name: "NO_COLOR unset but TERM=dumb",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("NO_COLOR", "")
				t.Setenv("TERM", "dumb")
			},
			want: false,
		},
		{
			name: "NO_COLOR and TERM unset, non-terminal",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("NO_COLOR", "")
				t.Setenv("TERM", "")
			},
			want: false,
		},
		{
			name: "NO_COLOR unset, TERM=xterm",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("NO_COLOR", "")
				t.Setenv("TERM", "xterm")
			},
			want: false, // Will be false in test environment as stdout is not a terminal.
		},
		{
			name: "NO_COLOR unset, TERM=xterm-256color",
			setup: func(t *testing.T) {
				t.Helper()

				t.Setenv("NO_COLOR", "")
				t.Setenv("TERM", "xterm-256color")
			},
			want: false, // Will be false in test environment as stdout is not a terminal.
		},
	}

	for _, tt := range tests { //nolint:paralleltest // t.Setenv is not thread-safe
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			got := cmdkit.WantColor()
			if got != tt.want {
				t.Errorf("WantColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadEnv(t *testing.T) { //nolint:paralleltest // t.Setenv is not thread-safe
	tests := []struct {
		name       string
		givePrefix string
		setupFlags func(fs *flag.FlagSet)
		setupEnv   func(t *testing.T)
		setFlags   func(fs *flag.FlagSet)
		want       map[string]string
		wantErr    bool
	}{
		{
			name:       "basic string flag",
			givePrefix: "app",
			setupFlags: func(fs *flag.FlagSet) {
				fs.String("config", "default.conf", "config file path")
			},
			setupEnv: func(t *testing.T) {
				t.Helper()

				t.Setenv("APP_CONFIG", "test.conf")
			},
			want: map[string]string{
				"config": "test.conf",
			},
		},
		{
			name:       "multiple flag types",
			givePrefix: "app",
			setupFlags: func(fs *flag.FlagSet) {
				fs.String("name", "default", "name value")
				fs.Int("count", 0, "count value")
				fs.Bool("verbose", false, "verbose output")
			},
			setupEnv: func(t *testing.T) {
				t.Helper()

				t.Setenv("APP_NAME", "test")
				t.Setenv("APP_COUNT", "42")
				t.Setenv("APP_VERBOSE", "true")
			},
			want: map[string]string{
				"name":    "test",
				"count":   "42",
				"verbose": "true",
			},
		},
		{
			name:       "cli flags take precedence",
			givePrefix: "app",
			setupFlags: func(fs *flag.FlagSet) {
				fs.String("name", "default", "name value")
			},
			setupEnv: func(t *testing.T) {
				t.Helper()

				t.Setenv("APP_NAME", "from_env")
			},
			setFlags: func(fs *flag.FlagSet) {
				fs.Set("name", "from_cli") //nolint:errcheck // ignore for tests
			},
			want: map[string]string{
				"name": "from_cli",
			},
		},
		{
			name:       "invalid env value",
			givePrefix: "app",
			setupFlags: func(fs *flag.FlagSet) {
				fs.Int("count", 0, "count value")
			},
			setupEnv: func(t *testing.T) {
				t.Helper()

				t.Setenv("APP_COUNT", "not_a_number")
			},
			wantErr: true,
		},
		{
			name:       "missing env variables",
			givePrefix: "app",
			setupFlags: func(fs *flag.FlagSet) {
				fs.String("name", "default", "name value")
			},
			setupEnv: func(_ *testing.T) {
				// No env variables set.
			},
			want: map[string]string{
				"name": "default",
			},
		},
		{
			name:       "different prefix",
			givePrefix: "myapp",
			setupFlags: func(fs *flag.FlagSet) {
				fs.String("config", "default.conf", "config file path")
			},
			setupEnv: func(t *testing.T) {
				t.Helper()

				t.Setenv("MYAPP_CONFIG", "test.conf")
			},
			want: map[string]string{
				"config": "test.conf",
			},
		},
		{
			name:       "dash in flag name",
			givePrefix: "app",
			setupFlags: func(fs *flag.FlagSet) {
				fs.String("config-file", "default.conf", "config file path")
			},
			setupEnv: func(t *testing.T) {
				t.Helper()

				t.Setenv("APP_CONFIG_FILE", "test.conf")
			},
			want: map[string]string{
				"config-file": "test.conf",
			},
		},
		{
			name:       "mixed case handling",
			givePrefix: "APP",
			setupFlags: func(fs *flag.FlagSet) {
				fs.String("ConfigFile", "default.conf", "config file path")
			},
			setupEnv: func(t *testing.T) {
				t.Helper()

				t.Setenv("APP_CONFIGFILE", "test.conf")
			},
			want: map[string]string{
				"ConfigFile": "test.conf",
			},
		},
		{
			name:       "empty prefix",
			givePrefix: "",
			setupFlags: func(fs *flag.FlagSet) {
				fs.String("name", "default", "name value")
			},
			setupEnv: func(t *testing.T) {
				t.Helper()

				t.Setenv("NAME", "test")
			},
			want: map[string]string{
				"name": "default",
			},
		},
		{
			name:       "multiple errors",
			givePrefix: "app",
			setupFlags: func(fs *flag.FlagSet) {
				fs.Int("count1", 0, "first count")
				fs.Int("count2", 0, "second count")
			},
			setupEnv: func(t *testing.T) {
				t.Helper()

				t.Setenv("APP_COUNT1", "not_a_number")
				t.Setenv("APP_COUNT2", "also_not_a_number")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests { //nolint:paralleltest // t.Setenv is not thread-safe
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet(t.Name(), flag.ContinueOnError)

			tt.setupFlags(fs)
			tt.setupEnv(t)

			if tt.setFlags != nil {
				tt.setFlags(fs)
			}

			err := cmdkit.LoadEnv(fs, tt.givePrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadEnv() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr {
				return
			}

			fs.VisitAll(func(f *flag.Flag) {
				if want, ok := tt.want[f.Name]; ok {
					if got := f.Value.String(); got != want {
						t.Errorf("LoadEnv() flag %q = %q, want %q", f.Name, got, want)
					}
				}
			})
		})
	}
}

func TestLoadEnv_RealWorld(t *testing.T) {
	type config struct {
		LogLevel  string
		Port      int
		Debug     bool
		APIKey    string
		CacheSize int64
	}

	var (
		cfg = &config{}
		fs  = flag.NewFlagSet("myapp", flag.ContinueOnError)
	)

	fs.StringVar(&cfg.LogLevel, "log-level", "info", "logging level")
	fs.IntVar(&cfg.Port, "port", 8080, "server port")
	fs.BoolVar(&cfg.Debug, "debug", false, "enable debug mode")
	fs.StringVar(&cfg.APIKey, "api-key", "", "API key")
	fs.Int64Var(&cfg.CacheSize, "cache-size", 1024, "cache size in MB")

	t.Setenv("MYAPP_LOG_LEVEL", "debug")
	t.Setenv("MYAPP_PORT", "9090")
	t.Setenv("MYAPP_DEBUG", "true")
	t.Setenv("MYAPP_API_KEY", "secret123")
	t.Setenv("MYAPP_CACHE_SIZE", "2048")

	err := cmdkit.LoadEnv(fs, "myapp")
	if err != nil {
		t.Fatalf("LoadEnv() error = %v", err)
	}

	want := config{
		LogLevel:  "debug",
		Port:      9090,
		Debug:     true,
		APIKey:    "secret123",
		CacheSize: 2048,
	}

	if *cfg != want {
		t.Errorf("LoadEnv() = %+v, want %+v", *cfg, want)
	}
}
