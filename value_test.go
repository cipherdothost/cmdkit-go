// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit_test

import (
	"flag"
	"testing"

	"go.cipher.host/cmdkit"
)

func TestRepeatValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func() *cmdkit.RepeatValue
		giveValue string
		want      string
		wantSlice []string
	}{
		{
			name: "empty value",
			setup: func() *cmdkit.RepeatValue {
				return &cmdkit.RepeatValue{}
			},
			want:      "",
			wantSlice: make([]string, 0),
		},
		{
			name: "single value",
			setup: func() *cmdkit.RepeatValue {
				return &cmdkit.RepeatValue{}
			},
			giveValue: "test",
			want:      "test",
			wantSlice: []string{
				"test",
			},
		},
		{
			name: "multiple values",
			setup: func() *cmdkit.RepeatValue {
				rv := &cmdkit.RepeatValue{}

				if err := rv.Set("first"); err != nil {
					t.Fatal(err)
				}

				if err := rv.Set("second"); err != nil {
					t.Fatal(err)
				}

				if err := rv.Set("third"); err != nil {
					t.Fatal(err)
				}

				return rv
			},
			want: "first, second, third",
			wantSlice: []string{
				"first",
				"second",
				"third",
			},
		},
		{
			name: "value with comma",
			setup: func() *cmdkit.RepeatValue {
				return &cmdkit.RepeatValue{}
			},
			giveValue: "a,b",
			want:      "a,b",
			wantSlice: []string{
				"a,b",
			},
		},
		{
			name: "add to existing values",
			setup: func() *cmdkit.RepeatValue {
				rv := &cmdkit.RepeatValue{}

				if err := rv.Set("existing"); err != nil {
					t.Fatal(err)
				}

				return rv
			},
			giveValue: "new",
			want:      "existing, new",
			wantSlice: []string{
				"existing",
				"new",
			},
		},
		{
			name: "multiple values with spaces",
			setup: func() *cmdkit.RepeatValue {
				rv := &cmdkit.RepeatValue{}

				if err := rv.Set("hello world"); err != nil {
					t.Fatal(err)
				}

				if err := rv.Set("goodbye world"); err != nil {
					t.Fatal(err)
				}
				return rv
			},
			want: "hello world, goodbye world",
			wantSlice: []string{
				"hello world",
				"goodbye world",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rv := tt.setup()

			if tt.giveValue != "" {
				if err := rv.Set(tt.giveValue); err != nil {
					t.Errorf("RepeatValue.Set(%q) unexpected error: %v", tt.giveValue, err)
				}
			}

			if got := rv.String(); got != tt.want {
				t.Errorf("RepeatValue.String() = %q, want %q", got, tt.want)
			}

			gotSlice := []string(*rv)
			if len(gotSlice) != len(tt.wantSlice) {
				t.Errorf("RepeatValue() slice length = %d, want %d", len(gotSlice), len(tt.wantSlice))

				return
			}

			for i := range gotSlice {
				if gotSlice[i] != tt.wantSlice[i] {
					t.Errorf("RepeatValue() slice[%d] = %q, want %q", i, gotSlice[i], tt.wantSlice[i])
				}
			}
		})
	}
}

func TestRepeatValue_FlagIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		giveArgs []string
		want     []string
	}{
		{
			name:     "no flags",
			giveArgs: make([]string, 0),
			want:     nil,
		},
		{
			name: "single flag",
			giveArgs: []string{
				"-tag",
				"test",
			},
			want: []string{
				"test",
			},
		},
		{
			name: "multiple flags",
			giveArgs: []string{
				"-tag",
				"first",
				"-tag",
				"second",
				"-tag",
				"third",
			},
			want: []string{
				"first",
				"second",
				"third",
			},
		},
		{
			name: "mixed with other flags",
			giveArgs: []string{
				"-other",
				"value",
				"-tag",
				"test",
				"-flag",
			},
			want: []string{
				"test",
			},
		},
		{
			name: "empty value",
			giveArgs: []string{
				"-tag",
				"",
			},
			want: []string{
				"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet("test", flag.ContinueOnError)

			var rv cmdkit.RepeatValue

			fs.Var(&rv, "tag", "repeatable tag flag")
			fs.String("other", "", "other flag")
			fs.Bool("flag", false, "boolean flag")

			err := fs.Parse(tt.giveArgs)
			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}

			got := []string(rv)
			if len(got) != len(tt.want) {
				t.Errorf("RepeatValue() len(value) = %d, want %d", len(got), len(tt.want))

				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("RepeatValue() value[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
