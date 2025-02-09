// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit

import (
	"flag"
	"strings"
)

// RepeatValue implements [flag.Value] to support repeated flag options. This
// allows flags to be specified multiple times to build a list of values.
//
// [flag.Value]: https://pkg.go.dev/flag#Value
type RepeatValue []string

// Compile-time check that RepeatValue implements the flag.Value interface.
var _ flag.Value = (*RepeatValue)(nil)

// String satisfies the [flag.Value] interface.
//
// [flag.Value]: https://pkg.go.dev/flag#Value
func (r *RepeatValue) String() string {
	return strings.Join(*r, ", ")
}

// Set satisfies the [flag.Value] interface.
//
// [flag.Value]: https://pkg.go.dev/flag#Value
func (r *RepeatValue) Set(value string) error {
	*r = append(*r, value)

	return nil
}
