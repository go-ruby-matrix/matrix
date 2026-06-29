// Copyright (c) the go-ruby-matrix/matrix authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package matrix is a pure-Go (CGO=0) reimplementation of Ruby's `matrix`
// standard library: Matrix and Vector with MRI-faithful behaviour.
//
// Like MRI, arithmetic is exact wherever the input is exact. Entries flow
// through the Num numeric tower (Integer / Rational / Float), so
// Matrix#determinant of an Integer matrix stays an Integer and Matrix#inverse
// produces exact Rationals — `New([][]any{{1,2},{3,4}}).Inverse()` renders
// `Matrix[[(-2/1), (1/1)], [(3/2), (-1/2)]]`, exactly as
// `Matrix[[1,2],[3,4]].inverse` does under Ruby 4.0.
//
// The formatting (ToS / Inspect), the error classes
// (ErrDimensionMismatch / ErrNotRegular / ErrOperationNotDefined) and the
// numeric promotion rules all track MRI so a binding (rbgo) can present this as
// the genuine `Matrix` class.
package matrix

import (
	"strings"
)

// Matrix is an immutable r×c matrix of Num entries, stored row-major.
type Matrix struct {
	rows int
	cols int
	e    []Num // len == rows*cols, row-major
}

func (m *Matrix) at(i, j int) Num     { return m.e[i*m.cols+j] }
func (m *Matrix) set(i, j int, v Num) { m.e[i*m.cols+j] = v }

// --- Constructors ---------------------------------------------------------

// New builds a Matrix from rows of values (each value convertible via the Num
// tower: int kinds, *big.Int, *big.Rat, float kinds, or Num). All rows must
// have the same length; an empty input yields a 0×0 matrix. It mirrors
// `Matrix[[...],[...]]` / `Matrix.rows`.
func New(rows [][]any) (*Matrix, error) {
	r := len(rows)
	if r == 0 {
		return &Matrix{}, nil
	}
	c := len(rows[0])
	m := &Matrix{rows: r, cols: c, e: make([]Num, r*c)}
	for i, row := range rows {
		if len(row) != c {
			return nil, ErrDimensionMismatch
		}
		for j, v := range row {
			n, err := numFromAny(v)
			if err != nil {
				return nil, err
			}
			m.set(i, j, n)
		}
	}
	return m, nil
}

// newFromNums builds a Matrix from a ready row-major Num slice (internal helper;
// the caller guarantees len(e) == r*c).
func newFromNums(r, c int, e []Num) *Matrix { return &Matrix{rows: r, cols: c, e: e} }

// Build returns an r×c matrix whose (i,j) entry is fn(i,j), like
// `Matrix.build(r, c) { |i, j| ... }`.
func Build(r, c int, fn func(i, j int) any) (*Matrix, error) {
	if r < 0 || c < 0 {
		return nil, ErrDimensionMismatch
	}
	m := &Matrix{rows: r, cols: c, e: make([]Num, r*c)}
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			n, err := numFromAny(fn(i, j))
			if err != nil {
				return nil, err
			}
			m.set(i, j, n)
		}
	}
	return m, nil
}

// Identity returns the n×n identity matrix (Integer entries), like
// `Matrix.identity(n)`.
func Identity(n int) *Matrix {
	m := Zero(n, n)
	for i := 0; i < n; i++ {
		m.set(i, i, NewInt(1))
	}
	return m
}

// Zero returns an r×c matrix of Integer zeros, like `Matrix.zero(r, c)`.
func Zero(r, c int) *Matrix {
	e := make([]Num, r*c)
	for i := range e {
		e[i] = NewInt(0)
	}
	return &Matrix{rows: r, cols: c, e: e}
}

// Diagonal returns a square matrix with the given values on the diagonal and
// Integer zeros elsewhere, like `Matrix.diagonal(*values)`.
func Diagonal(values ...any) (*Matrix, error) {
	n := len(values)
	m := Zero(n, n)
	for i, v := range values {
		nv, err := numFromAny(v)
		if err != nil {
			return nil, err
		}
		m.set(i, i, nv)
	}
	return m, nil
}

// Scalar returns the n×n matrix with value on the diagonal and zeros elsewhere,
// like `Matrix.scalar(n, value)`.
func Scalar(n int, value any) (*Matrix, error) {
	nv, err := numFromAny(value)
	if err != nil {
		return nil, err
	}
	m := Zero(n, n)
	for i := 0; i < n; i++ {
		m.set(i, i, nv)
	}
	return m, nil
}

// RowVector returns a 1×n matrix from the values, like
// `Matrix.row_vector([...])`.
func RowVector(values []any) (*Matrix, error) { return New([][]any{values}) }

// ColumnVector returns an n×1 matrix from the values, like
// `Matrix.column_vector([...])`.
func ColumnVector(values []any) (*Matrix, error) {
	rows := make([][]any, len(values))
	for i, v := range values {
		rows[i] = []any{v}
	}
	return New(rows)
}

// Rows builds a Matrix treating each inner slice as a row. It is an alias of New
// matching `Matrix.rows(...)`.
func Rows(rows [][]any) (*Matrix, error) { return New(rows) }

// Columns builds a Matrix treating each inner slice as a column, like
// `Matrix.columns(...)`.
func Columns(cols [][]any) (*Matrix, error) {
	m, err := New(cols)
	if err != nil {
		return nil, err
	}
	return m.Transpose(), nil
}

// HStack returns the matrices joined left-to-right; all must have equal row
// counts, like `Matrix.hstack(*matrices)`.
func HStack(ms ...*Matrix) (*Matrix, error) {
	if len(ms) == 0 {
		return &Matrix{}, nil
	}
	r := ms[0].rows
	totalC := 0
	for _, m := range ms {
		if m.rows != r {
			return nil, ErrDimensionMismatch
		}
		totalC += m.cols
	}
	out := &Matrix{rows: r, cols: totalC, e: make([]Num, r*totalC)}
	for i := 0; i < r; i++ {
		col := 0
		for _, m := range ms {
			for j := 0; j < m.cols; j++ {
				out.set(i, col, m.at(i, j))
				col++
			}
		}
	}
	return out, nil
}

// VStack returns the matrices stacked top-to-bottom; all must have equal column
// counts, like `Matrix.vstack(*matrices)`.
func VStack(ms ...*Matrix) (*Matrix, error) {
	if len(ms) == 0 {
		return &Matrix{}, nil
	}
	c := ms[0].cols
	totalR := 0
	for _, m := range ms {
		if m.cols != c {
			return nil, ErrDimensionMismatch
		}
		totalR += m.rows
	}
	out := &Matrix{rows: totalR, cols: c, e: make([]Num, totalR*c)}
	row := 0
	for _, m := range ms {
		for i := 0; i < m.rows; i++ {
			for j := 0; j < c; j++ {
				out.set(row, j, m.at(i, j))
			}
			row++
		}
	}
	return out, nil
}

// --- Accessors ------------------------------------------------------------

// RowCount returns the number of rows.
func (m *Matrix) RowCount() int { return m.rows }

// ColumnCount returns the number of columns.
func (m *Matrix) ColumnCount() int { return m.cols }

// At returns the (i,j) entry. It returns false when the indices are out of
// range, matching `Matrix#[]` which returns nil there.
func (m *Matrix) At(i, j int) (Num, bool) {
	if i < 0 || i >= m.rows || j < 0 || j >= m.cols {
		return Num{}, false
	}
	return m.at(i, j), true
}

// Row returns row i as a Vector, like `Matrix#row(i)`.
func (m *Matrix) Row(i int) (*Vector, bool) {
	if i < 0 || i >= m.rows {
		return nil, false
	}
	e := make([]Num, m.cols)
	copy(e, m.e[i*m.cols:(i+1)*m.cols])
	return &Vector{e: e}, true
}

// Column returns column j as a Vector, like `Matrix#column(j)`.
func (m *Matrix) Column(j int) (*Vector, bool) {
	if j < 0 || j >= m.cols {
		return nil, false
	}
	e := make([]Num, m.rows)
	for i := 0; i < m.rows; i++ {
		e[i] = m.at(i, j)
	}
	return &Vector{e: e}, true
}

// Each calls fn for every entry in row-major order, like `Matrix#each`.
func (m *Matrix) Each(fn func(v Num)) {
	for _, v := range m.e {
		fn(v)
	}
}

// EachWithIndex calls fn(v, i, j) for every entry in row-major order, like
// `Matrix#each_with_index`.
func (m *Matrix) EachWithIndex(fn func(v Num, i, j int)) {
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			fn(m.at(i, j), i, j)
		}
	}
}

// ToA returns the entries as a slice of rows (a fresh copy), like `Matrix#to_a`.
func (m *Matrix) ToA() [][]Num {
	out := make([][]Num, m.rows)
	for i := 0; i < m.rows; i++ {
		row := make([]Num, m.cols)
		copy(row, m.e[i*m.cols:(i+1)*m.cols])
		out[i] = row
	}
	return out
}

// Minor returns the submatrix spanning rows [r0,r1) and columns [c0,c1), like
// `Matrix#minor(r0...r1, c0...c1)`. Bounds are clamped to the matrix as MRI does
// for ranges that overshoot; an empty span yields an empty matrix.
func (m *Matrix) Minor(r0, r1, c0, c1 int) (*Matrix, error) {
	if r0 < 0 || c0 < 0 || r1 > m.rows || c1 > m.cols || r1 < r0 || c1 < c0 {
		return nil, ErrDimensionMismatch
	}
	nr, nc := r1-r0, c1-c0
	out := &Matrix{rows: nr, cols: nc, e: make([]Num, nr*nc)}
	for i := 0; i < nr; i++ {
		for j := 0; j < nc; j++ {
			out.set(i, j, m.at(r0+i, c0+j))
		}
	}
	return out, nil
}

// FirstMinor returns the matrix with row i and column j removed, like
// `Matrix#first_minor(i, j)`.
func (m *Matrix) FirstMinor(i, j int) (*Matrix, error) {
	if m.rows == 0 || m.cols == 0 {
		return nil, ErrDimensionMismatch
	}
	if i < 0 || i >= m.rows || j < 0 || j >= m.cols {
		return nil, ErrDimensionMismatch
	}
	nr, nc := m.rows-1, m.cols-1
	out := &Matrix{rows: nr, cols: nc, e: make([]Num, nr*nc)}
	ri := 0
	for r := 0; r < m.rows; r++ {
		if r == i {
			continue
		}
		cj := 0
		for c := 0; c < m.cols; c++ {
			if c == j {
				continue
			}
			out.set(ri, cj, m.at(r, c))
			cj++
		}
		ri++
	}
	return out, nil
}

// --- Predicates -----------------------------------------------------------

// Square reports whether the matrix is square, like `Matrix#square?`.
func (m *Matrix) Square() bool { return m.rows == m.cols }

// IsZero reports whether every entry is zero, like `Matrix#zero?`.
func (m *Matrix) IsZero() bool {
	for _, v := range m.e {
		if !v.IsZero() {
			return false
		}
	}
	return true
}

// IsDiagonal reports whether the matrix is square with zeros off the diagonal,
// like `Matrix#diagonal?`.
func (m *Matrix) IsDiagonal() bool {
	if !m.Square() {
		return false
	}
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			if i != j && !m.at(i, j).IsZero() {
				return false
			}
		}
	}
	return true
}

// Symmetric reports whether the matrix equals its transpose, like
// `Matrix#symmetric?`. Non-square matrices are not symmetric.
func (m *Matrix) Symmetric() bool {
	if !m.Square() {
		return false
	}
	for i := 0; i < m.rows; i++ {
		for j := i + 1; j < m.cols; j++ {
			if !m.at(i, j).Eql(m.at(j, i)) {
				return false
			}
		}
	}
	return true
}

// LowerTriangular reports whether all entries above the diagonal are zero, like
// `Matrix#lower_triangular?`.
func (m *Matrix) LowerTriangular() bool {
	for i := 0; i < m.rows; i++ {
		for j := i + 1; j < m.cols; j++ {
			if !m.at(i, j).IsZero() {
				return false
			}
		}
	}
	return true
}

// UpperTriangular reports whether all entries below the diagonal are zero, like
// `Matrix#upper_triangular?`.
func (m *Matrix) UpperTriangular() bool {
	for i := 0; i < m.rows; i++ {
		for j := 0; j < i && j < m.cols; j++ {
			if !m.at(i, j).IsZero() {
				return false
			}
		}
	}
	return true
}

// Singular reports whether the matrix is square and has determinant zero, like
// `Matrix#singular?`.
func (m *Matrix) Singular() (bool, error) {
	reg, err := m.Regular()
	if err != nil {
		return false, err
	}
	return !reg, nil
}

// Regular reports whether the matrix is square and non-singular (invertible),
// like `Matrix#regular?`.
func (m *Matrix) Regular() (bool, error) {
	if !m.Square() {
		return false, ErrDimensionMismatch
	}
	// On a square matrix Determinant never errors, so the error is discarded.
	d, _ := m.Determinant()
	return !d.IsZero(), nil
}

// Orthogonal reports whether the matrix is square and its transpose is its
// inverse (mᵀ·m == I), like `Matrix#orthogonal?`.
func (m *Matrix) Orthogonal() (bool, error) {
	if !m.Square() {
		return false, nil
	}
	// mᵀ has the same column count as m has rows, so on a square matrix the
	// product is always defined and Mul never errors.
	prod, _ := m.Transpose().Mul(m)
	return prod.Eql(Identity(m.rows)), nil
}

// --- Formatting -----------------------------------------------------------

// ToS renders the matrix as `Matrix[[...], [...]]`, identical to MRI's
// `Matrix#to_s` and `#inspect` for populated matrices. An empty matrix renders
// as `Matrix.empty(r, c)`.
func (m *Matrix) ToS() string {
	if m.rows == 0 || m.cols == 0 {
		var b strings.Builder
		b.WriteString("Matrix.empty(")
		b.WriteString(itoa(m.rows))
		b.WriteString(", ")
		b.WriteString(itoa(m.cols))
		b.WriteByte(')')
		return b.String()
	}
	var b strings.Builder
	b.WriteString("Matrix[")
	for i := 0; i < m.rows; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteByte('[')
		for j := 0; j < m.cols; j++ {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(m.at(i, j).String())
		}
		b.WriteByte(']')
	}
	b.WriteByte(']')
	return b.String()
}

// Inspect is an alias of ToS, matching that MRI's Matrix#inspect and #to_s
// produce the same text.
func (m *Matrix) Inspect() string { return m.ToS() }

// itoa formats a non-negative small int without importing strconv at call
// sites.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
