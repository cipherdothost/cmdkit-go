// SPDX-FileCopyrightText: 2009 The Go Authors
// SPDX-FileCopyrightText: 2024 Martin Tournoij
//
// SPDX-License-Identifier: LicenseRef-GoLicense

//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !zos && !windows && !solaris && !plan9

package term

func isTerminal(_ int) bool {
	return false
}
