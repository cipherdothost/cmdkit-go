// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit

import (
	"errors"
	"flag"
	"os"
)

// Exit is a convenience function that calls os.Exit with the error code
// associated with the error.

// Exit provides a consistent way to terminate a command-line application
// while respecting Unix exit code conventions.
//
// It handles the following cases:
//
// - nil error -> exit 0 (success)
// - flag.ErrHelp -> exit 0 (help display is not an error)
// - ExitError -> use provided exit code
// - other errors -> exit 1 (general failure)
//
// This function should typically be called at the top level of main().
func Exit(err error) {
	if err == nil {
		os.Exit(0)
	}

	if errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	}

	var e *ExitError
	if errors.As(err, &e) {
		os.Exit(e.Code())
	}

	os.Exit(1)
}
