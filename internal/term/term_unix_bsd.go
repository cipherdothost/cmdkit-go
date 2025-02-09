// SPDX-FileCopyrightText: 2009 The Go Authors
// SPDX-FileCopyrightText: 2024 Martin Tournoij
//
// SPDX-License-Identifier: LicenseRef-GoLicense

//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package term

import "syscall"

const (
	ioctlReadTermios = syscall.TIOCGETA
)
