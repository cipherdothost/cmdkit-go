<!--
SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>

SPDX-License-Identifier: CC0-1.0
-->

# term

This is a fork of [golang.org/x/term
v0.28.0](https://cs.opensource.google/go/x/term/+/refs/tags/v0.28.0:).

This fork makes some small style changes, update tests, removes the
dependency on [golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys),
removes support for Windows, and removes all code not needed by
[cmdkit](https://github.com/cipherdothost/cmdkit-go). It basically
strips the package down to the `IsTerminal` function.

All credits and copyright for the removal of the dependency on
[golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys) belong to
[Martin Tournoij](https://github.com/arp242), as we mostly just copied
part of their work from [their
fork](https://github.com/arp242/zli/tree/master/internal/term).
