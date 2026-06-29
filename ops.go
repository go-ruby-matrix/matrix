// Copyright (c) the go-ruby-matrix/matrix authors
//
// SPDX-License-Identifier: BSD-3-Clause

package matrix

// --- Element-wise and structural operations -------------------------------

// Add returns m + other (entry-wise); shapes must match, like `Matrix#+`.
func (m *Matrix) Add(other *Matrix) (*Matrix, error) {
	if m.rows != other.rows || m.cols != other.cols {
		return nil, ErrDimensionMismatch
	}
	e := make([]Num, len(m.e))
	for i := range m.e {
		e[i] = m.e[i].Add(other.e[i])
	}
	return newFromNums(m.rows, m.cols, e), nil
}

// Sub returns m - other (entry-wise); shapes must match, like `Matrix#-`.
func (m *Matrix) Sub(other *Matrix) (*Matrix, error) {
	if m.rows != other.rows || m.cols != other.cols {
		return nil, ErrDimensionMismatch
	}
	e := make([]Num, len(m.e))
	for i := range m.e {
		e[i] = m.e[i].Sub(other.e[i])
	}
	return newFromNums(m.rows, m.cols, e), nil
}

// Neg returns -m, like `Matrix#-@`.
func (m *Matrix) Neg() *Matrix {
	e := make([]Num, len(m.e))
	for i := range m.e {
		e[i] = m.e[i].Neg()
	}
	return newFromNums(m.rows, m.cols, e)
}

// Mul returns the matrix product m·other; m.cols must equal other.rows, like
// `Matrix#*` between matrices.
func (m *Matrix) Mul(other *Matrix) (*Matrix, error) {
	if m.cols != other.rows {
		return nil, ErrDimensionMismatch
	}
	out := &Matrix{rows: m.rows, cols: other.cols, e: make([]Num, m.rows*other.cols)}
	for i := 0; i < m.rows; i++ {
		for j := 0; j < other.cols; j++ {
			sum := NewInt(0)
			for k := 0; k < m.cols; k++ {
				sum = sum.Add(m.at(i, k).Mul(other.at(k, j)))
			}
			out.set(i, j, sum)
		}
	}
	return out, nil
}

// MulScalar returns m scaled by a single value, like `Matrix#*` with a scalar.
func (m *Matrix) MulScalar(scalar any) (*Matrix, error) {
	s, err := numFromAny(scalar)
	if err != nil {
		return nil, err
	}
	e := make([]Num, len(m.e))
	for i := range m.e {
		e[i] = m.e[i].Mul(s)
	}
	return newFromNums(m.rows, m.cols, e), nil
}

// MulVector returns the matrix-times-column-vector product m·v, returning a
// Vector, like `Matrix#*` with a Vector. m.cols must equal v.Size().
func (m *Matrix) MulVector(v *Vector) (*Vector, error) {
	if m.cols != len(v.e) {
		return nil, ErrDimensionMismatch
	}
	out := make([]Num, m.rows)
	for i := 0; i < m.rows; i++ {
		sum := NewInt(0)
		for k := 0; k < m.cols; k++ {
			sum = sum.Add(m.at(i, k).Mul(v.e[k]))
		}
		out[i] = sum
	}
	return &Vector{e: out}, nil
}

// Div returns m·other⁻¹ (matrix division), like `Matrix#/` between matrices.
func (m *Matrix) Div(other *Matrix) (*Matrix, error) {
	inv, err := other.Inverse()
	if err != nil {
		return nil, err
	}
	return m.Mul(inv)
}

// DivScalar returns m with each entry divided by scalar (exact: integers become
// Rationals), like `Matrix#/` with a scalar.
func (m *Matrix) DivScalar(scalar any) (*Matrix, error) {
	s, err := numFromAny(scalar)
	if err != nil {
		return nil, err
	}
	e := make([]Num, len(m.e))
	for i := range m.e {
		e[i] = m.e[i].Quo(s)
	}
	return newFromNums(m.rows, m.cols, e), nil
}

// Transpose returns the transpose mᵀ, like `Matrix#transpose` / `#t`.
func (m *Matrix) Transpose() *Matrix {
	out := &Matrix{rows: m.cols, cols: m.rows, e: make([]Num, len(m.e))}
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			out.set(j, i, m.at(i, j))
		}
	}
	return out
}

// Trace returns the sum of the diagonal entries; the matrix must be square,
// like `Matrix#trace` / `#tr`.
func (m *Matrix) Trace() (Num, error) {
	if !m.Square() {
		return Num{}, ErrDimensionMismatch
	}
	sum := NewInt(0)
	for i := 0; i < m.rows; i++ {
		sum = sum.Add(m.at(i, i))
	}
	return sum, nil
}

// Pow returns m raised to an integer power. Positive powers multiply m by
// itself; m**0 is the identity; negative powers use the inverse. The matrix must
// be square. Mirrors `Matrix#**`.
func (m *Matrix) Pow(p int) (*Matrix, error) {
	if !m.Square() {
		return nil, ErrDimensionMismatch
	}
	if p < 0 {
		inv, err := m.Inverse()
		if err != nil {
			return nil, err
		}
		return inv.Pow(-p)
	}
	// Square-and-multiply. Every product here is square×square of equal size, so
	// Mul never errors and the errors are discarded.
	result := Identity(m.rows)
	base := m
	for e := p; e > 0; e >>= 1 {
		if e&1 == 1 {
			result, _ = result.Mul(base)
		}
		if e > 1 {
			base, _ = base.Mul(base)
		}
	}
	return result, nil
}

// RoundEntries returns the matrix with every Float entry rounded to ndigits
// decimal places (Integer/Rational entries are unchanged), like
// `Matrix#round(ndigits)`.
func (m *Matrix) RoundEntries(ndigits int) *Matrix {
	e := make([]Num, len(m.e))
	for i := range m.e {
		e[i] = m.e[i].Round(ndigits)
	}
	return newFromNums(m.rows, m.cols, e)
}

// --- Determinant / inverse / rank -----------------------------------------

// Determinant returns the determinant; the matrix must be square, like
// `Matrix#determinant` / `#det`. It uses cofactor (Laplace) expansion with only
// +, − and ×, so an Integer matrix yields an Integer determinant exactly as MRI
// does (no Rational/Float drift).
func (m *Matrix) Determinant() (Num, error) {
	if !m.Square() {
		return Num{}, ErrDimensionMismatch
	}
	return det(m.ToA()), nil
}

// det computes the determinant of a square slice of Num via Laplace expansion
// along the first row. The 0×0 determinant is 1 (the empty product), matching
// MRI's Matrix.empty(0,0).det.
func det(a [][]Num) Num {
	n := len(a)
	switch n {
	case 0:
		return NewInt(1)
	case 1:
		return a[0][0]
	case 2:
		return a[0][0].Mul(a[1][1]).Sub(a[0][1].Mul(a[1][0]))
	}
	sum := NewInt(0)
	for j := 0; j < n; j++ {
		minor := make([][]Num, n-1)
		for i := 1; i < n; i++ {
			row := make([]Num, 0, n-1)
			for k := 0; k < n; k++ {
				if k != j {
					row = append(row, a[i][k])
				}
			}
			minor[i-1] = row
		}
		term := a[0][j].Mul(det(minor))
		if j&1 == 0 {
			sum = sum.Add(term)
		} else {
			sum = sum.Sub(term)
		}
	}
	return sum
}

// Inverse returns the inverse matrix; the matrix must be square and regular,
// like `Matrix#inverse` / `#inv`. It reproduces MRI's `inverse_from`
// algorithm step for step — Gauss-Jordan with partial pivoting on the largest
// absolute pivot, dividing with the exact Num.Quo — so an Integer matrix
// inverts to exact Rationals (`[[1,2],[3,4]].inverse` →
// `[[(-2/1),(1/1)],[(3/2),(-1/2)]]`) and a Float matrix reproduces MRI's exact
// rounding, including the same elimination order that yields values like
// `-1.9999999999999998`.
func (m *Matrix) Inverse() (*Matrix, error) {
	if !m.Square() {
		return nil, ErrDimensionMismatch
	}
	n := m.rows
	a := m.ToA()             // working copy of the source
	inv := Identity(n).ToA() // the augmented identity, evolved into the inverse

	for k := 0; k < n; k++ {
		// Partial pivot: pick the row with the largest |a[j][k]|.
		i := k
		akkAbs := absFloat(a[k][k])
		for j := k + 1; j < n; j++ {
			if v := absFloat(a[j][k]); v > akkAbs {
				i = j
				akkAbs = v
			}
		}
		if akkAbs == 0 {
			return nil, ErrNotRegular
		}
		if i != k {
			a[i], a[k] = a[k], a[i]
			inv[i], inv[k] = inv[k], inv[i]
		}
		akk := a[k][k]

		for ii := 0; ii < n; ii++ {
			if ii == k {
				continue
			}
			q := a[ii][k].Quo(akk)
			a[ii][k] = NewInt(0)
			for j := k + 1; j < n; j++ {
				a[ii][j] = a[ii][j].Sub(a[k][j].Mul(q))
			}
			for j := 0; j < n; j++ {
				inv[ii][j] = inv[ii][j].Sub(inv[k][j].Mul(q))
			}
		}
		for j := k + 1; j < n; j++ {
			a[k][j] = a[k][j].Quo(akk)
		}
		for j := 0; j < n; j++ {
			inv[k][j] = inv[k][j].Quo(akk)
		}
	}

	out := &Matrix{rows: n, cols: n, e: make([]Num, n*n)}
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			out.set(r, c, inv[r][c])
		}
	}
	return out, nil
}

// absFloat returns |n| as a float64, used only to choose the pivot (mirroring
// MRI's `a[j][k].abs` comparisons).
func absFloat(n Num) float64 {
	f := n.asFloat()
	if f < 0 {
		return -f
	}
	return f
}

// Rank returns the rank of the matrix (the number of linearly independent rows)
// computed by exact Gaussian elimination over the numeric tower, like
// `Matrix#rank`. It works for non-square matrices.
func (m *Matrix) Rank() int {
	if m.rows == 0 || m.cols == 0 {
		return 0
	}
	a := m.ToA()
	rows, cols := m.rows, m.cols
	rank := 0
	for col := 0; col < cols && rank < rows; col++ {
		pivot := -1
		for r := rank; r < rows; r++ {
			if !a[r][col].IsZero() {
				pivot = r
				break
			}
		}
		if pivot == -1 {
			continue
		}
		a[rank], a[pivot] = a[pivot], a[rank]
		pv := a[rank][col]
		for r := 0; r < rows; r++ {
			if r == rank || a[r][col].IsZero() {
				continue
			}
			factor := a[r][col].Quo(pv)
			for j := col; j < cols; j++ {
				a[r][j] = a[r][j].Sub(factor.Mul(a[rank][j]))
			}
		}
		rank++
	}
	return rank
}

// --- Equality / hash ------------------------------------------------------

// Eql reports whether two matrices have the same shape and numerically equal
// entries, like `Matrix#==`.
func (m *Matrix) Eql(other *Matrix) bool {
	if m.rows != other.rows || m.cols != other.cols {
		return false
	}
	for i := range m.e {
		if !m.e[i].Eql(other.e[i]) {
			return false
		}
	}
	return true
}

// Hash returns a hash of the matrix derived from its shape and entry values,
// consistent with Eql, like `Matrix#hash`.
func (m *Matrix) Hash() uint64 {
	h := uint64(1469598103934665603) // FNV-1a offset
	mix := func(s string) {
		for i := 0; i < len(s); i++ {
			h ^= uint64(s[i])
			h *= 1099511628211
		}
	}
	mix(itoa(m.rows))
	mix(itoa(m.cols))
	for _, v := range m.e {
		// Use the float value so that 2, (2/1) and 2.0 hash alike (consistent
		// with Eql treating them as equal).
		mix(formatRubyFloat(v.asFloat()))
	}
	return h
}
