// SPDX-FileCopyrightText: 2009 The Go Authors
// SPDX-FileCopyrightText: 2024 Martin Tournoij
//
// SPDX-License-Identifier: LicenseRef-GoLicense

package term

import (
	"syscall"
)

func isTerminal(fd int) bool {
	path, err := syscall.Fd2path(fd)
	if err != nil {
		return false
	}

	return path == "/dev/cons" || path == "/mnt/term/dev/cons"
}
