// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

// Package xflag provides a extension to the flag package.
package xflag

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Shorthand represents a mapping between a long flag name and its short form.
// This enables Unix-style short flags (e.g., "-h" for "--help") while
// maintaining the clearer long forms for scripts and documentation.
//
// The type ensures that short and long flags remain synchronized by sharing the
// same underlying flag.Value, preventing inconsistencies in flag handling.
type Shorthand struct {
	// Value is the flag's underlying value, shared between short and long
	// forms.
	Value flag.Value

	// Name is the full flag name, e.g., "help".
	Name string

	// Shorthand is the single-letter short form, e.g., "h".
	Shorthand string
}

// PrintFlags outputs formatted flag documentation that groups short and long
// forms together while maintaining consistent spacing and structure.
func PrintFlags(w io.Writer, fs *flag.FlagSet, shorthands []Shorthand) {
	if fs == nil {
		return
	}

	var (
		shorthandMap = make(map[string]string, len(shorthands))
		seenValues   = make(map[flag.Value]bool)
	)

	for _, sh := range shorthands {
		shorthandMap[sh.Name] = sh.Shorthand
		seenValues[sh.Value] = true
	}

	var maxWidth int

	fs.VisitAll(func(f *flag.Flag) {
		width := 2

		if sh, ok := shorthandMap[f.Name]; ok {
			width += len(sh) + 5
		}

		width += len(f.Name)

		if width > maxWidth {
			maxWidth = width
		}
	})

	fs.VisitAll(func(f *flag.Flag) {
		if sh, ok := shorthandMap[f.Name]; ok {
			printFlagWithShorthand(w, f, sh, maxWidth)
		} else if !seenValues[f.Value] {
			printSingleFlag(w, f, maxWidth)
		}
	})
}

func printFlagWithShorthand(w io.Writer, f *flag.Flag, shorthand string, width int) {
	var (
		prefix  = fmt.Sprintf("  -%s, --%s", shorthand, f.Name)
		padding = strings.Repeat(" ", width-len(prefix)+2)
	)

	fmt.Fprintf(w, "%s%s%s\n", prefix, padding, formatFlagUsage(f))
}

func printSingleFlag(w io.Writer, f *flag.Flag, width int) {
	var (
		prefix  = "  -" + f.Name
		padding = strings.Repeat(" ", width-len(prefix)+2)
	)

	fmt.Fprintf(w, "%s%s%s\n", prefix, padding, formatFlagUsage(f))
}

func formatFlagUsage(f *flag.Flag) string {
	var builder strings.Builder

	builder.WriteString(f.Usage)

	if f.DefValue != "" {
		builder.WriteString(" (defaults to ")
		builder.WriteString(strconv.Quote(f.DefValue))
		builder.WriteString(")")
	}

	return builder.String()
}
