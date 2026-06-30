// Copyright (c) the go-ruby-matrix/matrix authors
//
// SPDX-License-Identifier: BSD-3-Clause

package matrix

import (
	"fmt"
	"math"
	"strings"
)

// Vector is an immutable sequence of Num entries, the analogue of Ruby's
// Vector. Like Matrix, arithmetic is exact where the input is exact; magnitude,
// normalisation and angle go through Float because they involve a square root.
type Vector struct {
	e []Num
}

// NewVector builds a Vector from values (each convertible via the Num tower),
// like `Vector[...]` / `Vector.elements([...])`.
func NewVector(values []any) (*Vector, error) {
	e := make([]Num, len(values))
	for i, v := range values {
		n, err := numFromAny(v)
		if err != nil {
			return nil, err
		}
		e[i] = n
	}
	return &Vector{e: e}, nil
}

// Elements returns a copy of the entries, like `Vector#elements` / `#to_a`.
func (v *Vector) Elements() []Num {
	out := make([]Num, len(v.e))
	copy(out, v.e)
	return out
}

// Size returns the number of entries, like `Vector#size`.
func (v *Vector) Size() int { return len(v.e) }

// At returns the i-th entry, like `Vector#[]`. The second result is false when i
// is out of range (Ruby returns nil there).
func (v *Vector) At(i int) (Num, bool) {
	if i < 0 || i >= len(v.e) {
		return Num{}, false
	}
	return v.e[i], true
}

// Each calls fn for every entry in order, like `Vector#each`.
func (v *Vector) Each(fn func(n Num)) {
	for _, n := range v.e {
		fn(n)
	}
}

// Map returns a new Vector with fn applied to each entry, like `Vector#map`.
func (v *Vector) Map(fn func(n Num) any) (*Vector, error) {
	e := make([]Num, len(v.e))
	for i, n := range v.e {
		r, err := numFromAny(fn(n))
		if err != nil {
			return nil, err
		}
		e[i] = r
	}
	return &Vector{e: e}, nil
}

// Add returns v + other (entry-wise); sizes must match, like `Vector#+`.
func (v *Vector) Add(other *Vector) (*Vector, error) {
	if len(v.e) != len(other.e) {
		return nil, ErrDimensionMismatch
	}
	e := make([]Num, len(v.e))
	for i := range v.e {
		e[i] = v.e[i].Add(other.e[i])
	}
	return &Vector{e: e}, nil
}

// Sub returns v - other (entry-wise); sizes must match, like `Vector#-`.
func (v *Vector) Sub(other *Vector) (*Vector, error) {
	if len(v.e) != len(other.e) {
		return nil, ErrDimensionMismatch
	}
	e := make([]Num, len(v.e))
	for i := range v.e {
		e[i] = v.e[i].Sub(other.e[i])
	}
	return &Vector{e: e}, nil
}

// Mul returns v scaled by a single value, like `Vector#*` with a scalar.
func (v *Vector) Mul(scalar any) (*Vector, error) {
	s, err := numFromAny(scalar)
	if err != nil {
		return nil, err
	}
	e := make([]Num, len(v.e))
	for i := range v.e {
		e[i] = v.e[i].Mul(s)
	}
	return &Vector{e: e}, nil
}

// InnerProduct returns the dot product v·other; sizes must match, like
// `Vector#inner_product` / `#dot`.
func (v *Vector) InnerProduct(other *Vector) (Num, error) {
	if len(v.e) != len(other.e) {
		return Num{}, ErrDimensionMismatch
	}
	sum := NewInt(0)
	for i := range v.e {
		sum = sum.Add(v.e[i].Mul(other.e[i]))
	}
	return sum, nil
}

// CrossProduct returns the cross product with one other vector, like the binary
// form of `Vector#cross_product`. It mirrors MRI exactly: the receiver must have
// dimension at least 2, and the cross product with a single argument is only
// defined when the receiver is 3-dimensional. MRI implements cross_product(*vs)
// requiring vs.size == size-2, so for a single argument any size other than 3
// is the wrong number of arguments and raises ArgumentError; a sub-2 dimension
// instead raises ErrOperationNotDefined, and a dimension mismatch between the
// two vectors raises ErrDimensionMismatch.
func (v *Vector) CrossProduct(other *Vector) (*Vector, error) {
	size := len(v.e)
	if size < 2 {
		return nil, ErrOperationNotDefined
	}
	// MRI: raise ArgumentError unless vs.size == size-2. Here vs.size == 1.
	if size-2 != 1 {
		return nil, argumentError(fmt.Sprintf("wrong number of arguments (1 for %d)", size-2))
	}
	if len(other.e) != size {
		return nil, ErrDimensionMismatch
	}
	a, b := v.e, other.e
	e := []Num{
		a[1].Mul(b[2]).Sub(a[2].Mul(b[1])),
		a[2].Mul(b[0]).Sub(a[0].Mul(b[2])),
		a[0].Mul(b[1]).Sub(a[1].Mul(b[0])),
	}
	return &Vector{e: e}, nil
}

// Magnitude returns the Euclidean norm √(Σ vᵢ²) as a Float, like
// `Vector#magnitude` / `#norm` / `#r`.
func (v *Vector) Magnitude() Num {
	sum := NewInt(0)
	for _, n := range v.e {
		sum = sum.Add(n.Mul(n))
	}
	return sum.Sqrt()
}

// Normalize returns v scaled to unit magnitude; the zero vector cannot be
// normalised, like `Vector#normalize`.
func (v *Vector) Normalize() (*Vector, error) {
	mag := v.Magnitude()
	if mag.IsZero() {
		return nil, ErrOperationNotDefined
	}
	e := make([]Num, len(v.e))
	for i := range v.e {
		e[i] = v.e[i].Quo(mag)
	}
	return &Vector{e: e}, nil
}

// Angle returns the angle in radians between v and other (a Float), like
// `Vector#angle_with`. Neither vector may be zero.
func (v *Vector) Angle(other *Vector) (Num, error) {
	if len(v.e) != len(other.e) {
		return Num{}, ErrDimensionMismatch
	}
	// Sizes match (checked above), so InnerProduct never errors here.
	dot, _ := v.InnerProduct(other)
	mv := v.Magnitude()
	mo := other.Magnitude()
	if mv.IsZero() || mo.IsZero() {
		return Num{}, ErrOperationNotDefined
	}
	cos := dot.asFloat() / (mv.asFloat() * mo.asFloat())
	// Clamp to [-1,1] to absorb rounding, as MRI does.
	if cos > 1 {
		cos = 1
	} else if cos < -1 {
		cos = -1
	}
	return NewFloat(math.Acos(cos)), nil
}

// Eql reports whether two vectors have the same size and numerically equal
// entries, like `Vector#==`.
func (v *Vector) Eql(other *Vector) bool {
	if len(v.e) != len(other.e) {
		return false
	}
	for i := range v.e {
		if !v.e[i].Eql(other.e[i]) {
			return false
		}
	}
	return true
}

// ToS renders the vector as `Vector[...]`, identical to MRI's `Vector#to_s` and
// `#inspect`.
func (v *Vector) ToS() string {
	var b strings.Builder
	b.WriteString("Vector[")
	for i, n := range v.e {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(n.String())
	}
	b.WriteByte(']')
	return b.String()
}

// Inspect is an alias of ToS.
func (v *Vector) Inspect() string { return v.ToS() }

// Independent reports whether the given vectors are linearly independent, like
// `Vector.independent?(*vs)`. Vectors of differing sizes are not independent
// (MRI raises ErrDimensionMismatch); this returns an error for that case.
func Independent(vs ...*Vector) (bool, error) {
	if len(vs) == 0 {
		return true, nil
	}
	size := len(vs[0].e)
	rows := make([][]any, len(vs))
	for i, vec := range vs {
		if len(vec.e) != size {
			return false, ErrDimensionMismatch
		}
		// Each vector becomes a row; rank == count ⇔ independent.
		row := make([]any, size)
		for j, n := range vec.e {
			row[j] = n
		}
		rows[i] = row
	}
	if len(vs) > size {
		return false, nil
	}
	// All rows have the same length (checked above), so New never errors.
	m, _ := New(rows)
	return m.Rank() == len(vs), nil
}
