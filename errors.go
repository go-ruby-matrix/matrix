// Copyright (c) the go-ruby-matrix/matrix authors
//
// SPDX-License-Identifier: BSD-3-Clause

package matrix

import "errors"

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
)
