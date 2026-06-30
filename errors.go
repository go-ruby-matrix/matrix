// Copyright (c) the go-ruby-matrix/matrix authors
//
// SPDX-License-Identifier: BSD-3-Clause

package matrix

import (
	"errors"
	"fmt"
)

// These mirror the exceptions Ruby's matrix library raises under
// ExceptionForMatrix. The messages match MRI's so a binding can surface them
// verbatim.
var (
	// ErrDimensionMismatch corresponds to Matrix::ErrDimensionMismatch
	// ("Dimension mismatch"): operands whose shapes are incompatible.
	ErrDimensionMismatch = errors.New("Dimension mismatch")

	// ErrNotRegular corresponds to Matrix::ErrNotRegular ("Not Regular
	// Matrix"): inverting or dividing by a singular matrix.
	ErrNotRegular = errors.New("Not Regular Matrix")

	// ErrOperationNotDefined corresponds to Matrix::ErrOperationNotDefined: an
	// operation that is not defined for the given operand.
	ErrOperationNotDefined = errors.New("operation not defined")

	// ErrArgument corresponds to Ruby's ArgumentError. The concrete error
	// carries MRI's exact message (e.g. "wrong number of arguments (1 for 0)")
	// and wraps this sentinel so callers can match it with errors.Is.
	ErrArgument = errors.New("ArgumentError")
)

// argumentError builds an ArgumentError-equivalent error whose Error() is the
// MRI message verbatim while still matching ErrArgument under errors.Is.
func argumentError(msg string) error {
	return fmt.Errorf("%s%w", msg, hidden{ErrArgument})
}

// hidden wraps a sentinel so errors.Is finds it without the sentinel's own text
// appearing in the formatted message.
type hidden struct{ err error }

func (h hidden) Error() string { return "" }
func (h hidden) Unwrap() error { return h.err }
