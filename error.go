// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit

// Error is an [immutable error] type.
//
// [immutable error]: https://dave.cheney.net/2016/04/07/constant-errors
type Error string

// Error implements the error interface for Error.
func (e Error) Error() string {
	return string(e)
}

// ExitError wraps an error with an exit code, allowing commands to control
// their process exit status in a Unix-compatible way similar to the [errno]
// standard in C.
//
// [errno]: https://linux.die.net/man/3/errno
type ExitError struct {
	// Err is the underlying error being wrapped.
	Err error

	// No is the process exit code to use. E.g., 0 for success, 1-255 for
	// errors.
	No int
}

// NewExitError returns a new Error instance with the given error code.
func NewExitError(err error, code int) *ExitError {
	return &ExitError{
		Err: err,
		No:  code,
	}
}

// Error implements the error interface.
func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return "undefined error"
}

// Code returns the error code.
func (e *ExitError) Code() int {
	return e.No
}

// Unwrap returns the underlying error.
func (e *ExitError) Unwrap() error {
	return e.Err
}
