// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit_test

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"testing"

	"go.cipher.host/cmdkit"
)

func TestExit_NilError(t *testing.T) {
	t.Parallel()

	if os.Getenv("TEST_EXIT_NILERROR") == "1" {
		cmdkit.Exit(nil)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExit_NilError") //nolint:gosec // this is not a security issue
	cmd.Env = append(os.Environ(), "TEST_EXIT_NILERROR=1")

	var exitErr *exec.ExitError
	if err := cmd.Run(); err != nil {
		if errors.As(err, &exitErr) && !exitErr.Success() {
			return
		}

		t.Fatalf("Exit() process ran with err = %v, want exit status 0", err)
	}
}

func TestExit_ErrHelp(t *testing.T) {
	t.Parallel()

	if os.Getenv("TEST_EXIT_ERRHELP") == "1" {
		cmdkit.Exit(flag.ErrHelp)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExit_ErrHelp") //nolint:gosec // this is not a security issue
	cmd.Env = append(os.Environ(), "TEST_EXIT_ERRHELP=1")

	var exitErr *exec.ExitError
	if err := cmd.Run(); err != nil {
		if errors.As(err, &exitErr) && !exitErr.Success() {
			return
		}

		t.Fatalf("Exit() process ran with err = %v, want exit status 0", err)
	}
}

func TestExit_ExitError(t *testing.T) {
	t.Parallel()

	if os.Getenv("TEST_EXIT_EXITERROR") == "1" {
		testErr := cmdkit.NewExitError(errGeneric, 1)

		cmdkit.Exit(testErr)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExit_ExitError") //nolint:gosec // this is not a security issue
	cmd.Env = append(os.Environ(), "TEST_EXIT_EXITERROR=1")

	var exitErr *exec.ExitError
	if err := cmd.Run(); err != nil {
		if errors.As(err, &exitErr) && !exitErr.Success() {
			return
		}

		t.Fatalf("Exit() process ran with err = %v, want exit status 0", err)
	}

	if exitErr.ExitCode() != 1 {
		t.Fatalf("Exit() process exited with code %d, want 1", exitErr.ExitCode())
	}
}

func TestExit_GenericError(t *testing.T) {
	t.Parallel()

	if os.Getenv("TEST_EXIT_GENERICERROR") == "1" {
		cmdkit.Exit(errGeneric)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExit_GenericError") //nolint:gosec // this is not a security issue
	cmd.Env = append(os.Environ(), "TEST_EXIT_GENERICERROR=1")

	var exitErr *exec.ExitError
	if err := cmd.Run(); err != nil {
		if errors.As(err, &exitErr) && !exitErr.Success() {
			return
		}

		t.Fatalf("Exit() process ran with err = %v, want exit status 0", err)
	}

	if exitErr.ExitCode() != 1 {
		t.Fatalf("Exit() process exited with code %d, want 1", exitErr.ExitCode())
	}
}
