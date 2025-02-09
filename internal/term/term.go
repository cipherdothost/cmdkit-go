// SPDX-FileCopyrightText: 2009 The Go Authors
// SPDX-FileCopyrightText: 2024 Martin Tournoij
//
// SPDX-License-Identifier: LicenseRef-GoLicense

// Package term provides support functions for dealing with terminals, as
// commonly found on UNIX systems.
package term

// IsTerminal returns whether the given file descriptor is a terminal.
func IsTerminal(fd int) bool {
	return isTerminal(fd)
}
