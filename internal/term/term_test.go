// SPDX-FileCopyrightText: 2009 The Go Authors
// SPDX-FileCopyrightText: 2024 Martin Tournoij
//
// SPDX-License-Identifier: LicenseRef-GoLicense

package term_test

import (
	"os"
	"runtime"
	"testing"

	"go.cipher.host/cmdkit/internal/term"
)

func TestIsTerminalTempFile(t *testing.T) {
	t.Parallel()

	file, err := os.CreateTemp(t.TempDir(), "TestIsTerminalTempFile")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.Remove(file.Name())
		file.Close()
	})

	if term.IsTerminal(int(file.Fd())) {
		t.Fatalf("IsTerminal unexpectedly returned true for temporary file %s", file.Name())
	}
}

func TestIsTerminalTerm(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "linux" {
		t.Skipf("unknown terminal path for GOOS %v", runtime.GOOS)
	}

	file, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		file.Close()
	})

	if !term.IsTerminal(int(file.Fd())) {
		t.Fatalf("IsTerminal unexpectedly returned false for terminal file %s", file.Name())
	}
}
