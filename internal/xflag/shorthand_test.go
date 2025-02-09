// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package xflag_test

import (
	"bytes"
	"flag"
	"testing"

	"go.cipher.host/cmdkit/internal/xflag"
)

func TestPrintFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFlags func(*flag.FlagSet) []xflag.Shorthand
		want       string
	}{
		{
			name: "single flag",
			setupFlags: func(fs *flag.FlagSet) []xflag.Shorthand {
				fs.String("name", "", "name value")
				return nil
			},
			want: "  -name name value\n",
		},
		{
			name: "flag with default value",
			setupFlags: func(fs *flag.FlagSet) []xflag.Shorthand {
				fs.String("name", "default", "name value")

				return nil
			},
			want: "  -name name value (defaults to \"default\")\n",
		},
		{
			name: "multiple flags",
			setupFlags: func(fs *flag.FlagSet) []xflag.Shorthand {
				fs.String("name", "", "name value")
				fs.Int("count", 0, "count value")
				fs.Bool("verbose", false, "verbose output")

				return nil
			},
			want: "  -count   count value (defaults to \"0\")\n" +
				"  -name    name value\n" +
				"  -verbose verbose output (defaults to \"false\")\n",
		},
		{
			name: "single shorthand",
			setupFlags: func(fs *flag.FlagSet) []xflag.Shorthand {
				var verbose bool

				fs.BoolVar(&verbose, "verbose", false, "verbose output")
				fs.BoolVar(&verbose, "v", false, "verbose output")

				return []xflag.Shorthand{
					{
						Name:      "verbose",
						Shorthand: "v",
						Value:     fs.Lookup("verbose").Value,
					},
				}
			},
			want: "  -v, --verbose  verbose output (defaults to \"false\")\n",
		},
		{
			name: "multiple shorthands",
			setupFlags: func(fs *flag.FlagSet) []xflag.Shorthand {
				var (
					verbose bool
					output  string
				)

				fs.BoolVar(&verbose, "verbose", false, "verbose output")
				fs.BoolVar(&verbose, "v", false, "verbose output")
				fs.StringVar(&output, "output", "", "output file")
				fs.StringVar(&output, "o", "", "output file")

				return []xflag.Shorthand{
					{
						Name:      "verbose",
						Shorthand: "v",
						Value:     fs.Lookup("verbose").Value,
					},
					{
						Name:      "output",
						Shorthand: "o",
						Value:     fs.Lookup("output").Value,
					},
				}
			},
			want: "  -o, --output   output file\n" +
				"  -v, --verbose  verbose output (defaults to \"false\")\n",
		},
		{
			name: "mixed regular and shorthand flags",
			setupFlags: func(fs *flag.FlagSet) []xflag.Shorthand {
				var (
					verbose bool
					name    string
					count   int
				)

				fs.BoolVar(&verbose, "verbose", false, "verbose output")
				fs.BoolVar(&verbose, "v", false, "verbose output")
				fs.StringVar(&name, "name", "", "name value")
				fs.IntVar(&count, "count", 0, "count value")

				return []xflag.Shorthand{
					{
						Name:      "verbose",
						Shorthand: "v",
						Value:     fs.Lookup("verbose").Value,
					},
				}
			},
			want: "  -count         count value (defaults to \"0\")\n" +
				"  -name          name value\n" +
				"  -v, --verbose  verbose output (defaults to \"false\")\n",
		},
		{
			name: "long flag names",
			setupFlags: func(fs *flag.FlagSet) []xflag.Shorthand {
				fs.String("very-long-flag-name", "", "flag with a very long name")
				fs.String("short", "", "short flag")

				return nil
			},
			want: "  -short               short flag\n" +
				"  -very-long-flag-name flag with a very long name\n",
		},
		{
			name: "all flag types",
			setupFlags: func(fs *flag.FlagSet) []xflag.Shorthand {
				fs.Bool("bool", false, "bool value")
				fs.Int("int", 0, "int value")
				fs.Int64("int64", 0, "int64 value")
				fs.Uint("uint", 0, "uint value")
				fs.Uint64("uint64", 0, "uint64 value")
				fs.Float64("float64", 0, "float64 value")
				fs.String("string", "", "string value")
				fs.Duration("duration", 0, "duration value")

				return nil
			},
			want: "  -bool     bool value (defaults to \"false\")\n" +
				"  -duration duration value (defaults to \"0s\")\n" +
				"  -float64  float64 value (defaults to \"0\")\n" +
				"  -int      int value (defaults to \"0\")\n" +
				"  -int64    int64 value (defaults to \"0\")\n" +
				"  -string   string value\n" +
				"  -uint     uint value (defaults to \"0\")\n" +
				"  -uint64   uint64 value (defaults to \"0\")\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				buf        bytes.Buffer
				fs         = flag.NewFlagSet("test", flag.ContinueOnError)
				shorthands = tt.setupFlags(fs)
			)

			xflag.PrintFlags(&buf, fs, shorthands)

			if got := buf.String(); got != tt.want {
				t.Errorf("PrintFlags() output = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("nil FlagSet", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		xflag.PrintFlags(&buf, nil, nil)

		if got := buf.String(); got != "" {
			t.Errorf("PrintFlags() with nil FlagSet wrote %q, want empty string", got)
		}
	})
}
