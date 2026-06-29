// Copyright (c) the go-ruby-matrix/matrix authors
//
// SPDX-License-Identifier: BSD-3-Clause

package matrix

import (
	"errors"
	"math"
	"math/big"
	"testing"
)

// mat is a test helper that builds a Matrix or fails the test.
func mat(t *testing.T, rows [][]any) *Matrix {
	t.Helper()
	m, err := New(rows)
	if err != nil {
		t.Fatalf("New(%v): %v", rows, err)
	}
	return m
}

// vec is a test helper that builds a Vector or fails the test.
func vec(t *testing.T, xs ...any) *Vector {
	t.Helper()
	v, err := NewVector(xs)
	if err != nil {
		t.Fatalf("NewVector(%v): %v", xs, err)
	}
	return v
}

func TestConstructors(t *testing.T) {
	if got := Identity(2).ToS(); got != "Matrix[[1, 0], [0, 1]]" {
		t.Errorf("Identity = %s", got)
	}
	if got := Zero(2, 3).ToS(); got != "Matrix[[0, 0, 0], [0, 0, 0]]" {
		t.Errorf("Zero = %s", got)
	}
	b, err := Build(2, 2, func(i, j int) any { return i*2 + j })
	if err != nil || b.ToS() != "Matrix[[0, 1], [2, 3]]" {
		t.Errorf("Build = %s, %v", b.ToS(), err)
	}
	d, err := Diagonal(1, 2, 3)
	if err != nil || d.ToS() != "Matrix[[1, 0, 0], [0, 2, 0], [0, 0, 3]]" {
		t.Errorf("Diagonal = %s, %v", d.ToS(), err)
	}
	s, err := Scalar(3, 5)
	if err != nil || s.ToS() != "Matrix[[5, 0, 0], [0, 5, 0], [0, 0, 5]]" {
		t.Errorf("Scalar = %s, %v", s.ToS(), err)
	}
	rv, err := RowVector([]any{1, 2, 3})
	if err != nil || rv.ToS() != "Matrix[[1, 2, 3]]" {
		t.Errorf("RowVector = %s, %v", rv.ToS(), err)
	}
	cv, err := ColumnVector([]any{1, 2, 3})
	if err != nil || cv.ToS() != "Matrix[[1], [2], [3]]" {
		t.Errorf("ColumnVector = %s, %v", cv.ToS(), err)
	}
	cols, err := Columns([][]any{{1, 2}, {3, 4}})
	if err != nil || cols.ToS() != "Matrix[[1, 3], [2, 4]]" {
		t.Errorf("Columns = %s, %v", cols.ToS(), err)
	}
	rows, err := Rows([][]any{{1, 2}, {3, 4}})
	if err != nil || rows.ToS() != "Matrix[[1, 2], [3, 4]]" {
		t.Errorf("Rows = %s, %v", rows.ToS(), err)
	}
}

func TestNewErrors(t *testing.T) {
	if _, err := New([][]any{{1, 2}, {3}}); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("ragged rows err = %v", err)
	}
	if _, err := New([][]any{{"x"}}); err == nil {
		t.Error("non-numeric entry: want error")
	}
	if m, err := New([][]any{}); err != nil || m.RowCount() != 0 {
		t.Errorf("empty New = %v, %v", m, err)
	}
}

func TestNumFromAnyKinds(t *testing.T) {
	bi := new(big.Int).SetUint64(uint64(math.MaxUint64))
	cases := []any{
		int(1), int8(2), int16(3), int32(4), int64(5),
		uint(6), uint64(7), bi, big.NewRat(1, 2),
		float32(1.5), float64(2.5), NewInt(9),
	}
	for _, c := range cases {
		if _, err := numFromAny(c); err != nil {
			t.Errorf("numFromAny(%T): %v", c, err)
		}
	}
	if _, err := numFromAny("nope"); err == nil {
		t.Error("string entry: want error")
	}
}

func TestBuildNegative(t *testing.T) {
	if _, err := Build(-1, 2, func(i, j int) any { return 0 }); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Build(-1,2) err = %v", err)
	}
	if _, err := Build(1, 1, func(i, j int) any { return "x" }); err == nil {
		t.Error("Build bad entry: want error")
	}
}

func TestConstructorErrors(t *testing.T) {
	if _, err := Diagonal("x"); err == nil {
		t.Error("Diagonal bad value: want error")
	}
	if _, err := Scalar(2, "x"); err == nil {
		t.Error("Scalar bad value: want error")
	}
	if _, err := RowVector([]any{"x"}); err == nil {
		t.Error("RowVector bad: want error")
	}
	if _, err := ColumnVector([]any{"x"}); err == nil {
		t.Error("ColumnVector bad: want error")
	}
	if _, err := Columns([][]any{{"x"}}); err == nil {
		t.Error("Columns bad: want error")
	}
}

func TestStack(t *testing.T) {
	h, err := HStack(mat(t, [][]any{{1}, {2}}), mat(t, [][]any{{3}, {4}}))
	if err != nil || h.ToS() != "Matrix[[1, 3], [2, 4]]" {
		t.Errorf("HStack = %s, %v", h.ToS(), err)
	}
	v, err := VStack(mat(t, [][]any{{1, 2}}), mat(t, [][]any{{3, 4}}))
	if err != nil || v.ToS() != "Matrix[[1, 2], [3, 4]]" {
		t.Errorf("VStack = %s, %v", v.ToS(), err)
	}
	if e, _ := HStack(); e.RowCount() != 0 {
		t.Error("HStack() empty")
	}
	if e, _ := VStack(); e.RowCount() != 0 {
		t.Error("VStack() empty")
	}
	if _, err := HStack(mat(t, [][]any{{1}}), mat(t, [][]any{{1}, {2}})); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("HStack mismatch = %v", err)
	}
	if _, err := VStack(mat(t, [][]any{{1}}), mat(t, [][]any{{1, 2}})); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("VStack mismatch = %v", err)
	}
}

func TestAccessors(t *testing.T) {
	m := mat(t, [][]any{{1, 2, 3}, {4, 5, 6}})
	if m.RowCount() != 2 || m.ColumnCount() != 3 {
		t.Error("counts")
	}
	if v, ok := m.At(0, 1); !ok || v.String() != "2" {
		t.Errorf("At = %s, %v", v, ok)
	}
	if _, ok := m.At(5, 0); ok {
		t.Error("At out of range should be !ok")
	}
	if _, ok := m.At(0, 9); ok {
		t.Error("At col out of range")
	}
	r, ok := m.Row(0)
	if !ok || r.ToS() != "Vector[1, 2, 3]" {
		t.Errorf("Row = %s", r.ToS())
	}
	if _, ok := m.Row(9); ok {
		t.Error("Row out of range")
	}
	c, ok := m.Column(1)
	if !ok || c.ToS() != "Vector[2, 5]" {
		t.Errorf("Column = %s", c.ToS())
	}
	if _, ok := m.Column(9); ok {
		t.Error("Column out of range")
	}
	var count int
	m.Each(func(v Num) { count++ })
	if count != 6 {
		t.Errorf("Each visited %d", count)
	}
	var sumIdx int
	m.EachWithIndex(func(v Num, i, j int) { sumIdx += i + j })
	if sumIdx != 9 {
		t.Errorf("EachWithIndex idx sum %d", sumIdx)
	}
	a := m.ToA()
	if len(a) != 2 || len(a[0]) != 3 || a[1][2].String() != "6" {
		t.Errorf("ToA = %v", a)
	}
}

func TestMinors(t *testing.T) {
	m := mat(t, [][]any{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}})
	mn, err := m.Minor(0, 2, 0, 2)
	if err != nil || mn.ToS() != "Matrix[[1, 2], [4, 5]]" {
		t.Errorf("Minor = %s, %v", mn.ToS(), err)
	}
	if _, err := m.Minor(0, 9, 0, 1); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Minor oob = %v", err)
	}
	if _, err := m.Minor(2, 1, 0, 1); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Minor reversed = %v", err)
	}
	fm, err := m.FirstMinor(0, 0)
	if err != nil || fm.ToS() != "Matrix[[5, 6], [8, 9]]" {
		t.Errorf("FirstMinor = %s, %v", fm.ToS(), err)
	}
	if _, err := m.FirstMinor(9, 0); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("FirstMinor oob = %v", err)
	}
	empty := Zero(0, 0)
	if _, err := empty.FirstMinor(0, 0); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("FirstMinor empty = %v", err)
	}
}

func TestArithmetic(t *testing.T) {
	a := mat(t, [][]any{{1, 2}, {3, 4}})
	b := mat(t, [][]any{{5, 6}, {7, 8}})
	sum, _ := a.Add(b)
	if sum.ToS() != "Matrix[[6, 8], [10, 12]]" {
		t.Errorf("Add = %s", sum.ToS())
	}
	diff, _ := a.Sub(b)
	if diff.ToS() != "Matrix[[-4, -4], [-4, -4]]" {
		t.Errorf("Sub = %s", diff.ToS())
	}
	if a.Neg().ToS() != "Matrix[[-1, -2], [-3, -4]]" {
		t.Errorf("Neg = %s", a.Neg().ToS())
	}
	prod, _ := a.Mul(Identity(2))
	if prod.ToS() != "Matrix[[1, 2], [3, 4]]" {
		t.Errorf("Mul = %s", prod.ToS())
	}
	sc, _ := a.MulScalar(2)
	if sc.ToS() != "Matrix[[2, 4], [6, 8]]" {
		t.Errorf("MulScalar = %s", sc.ToS())
	}
	mv, _ := a.MulVector(vec(t, 1, 1))
	if mv.ToS() != "Vector[3, 7]" {
		t.Errorf("MulVector = %s", mv.ToS())
	}
	dv, _ := a.DivScalar(2)
	if dv.ToS() != "Matrix[[(1/2), (1/1)], [(3/2), (2/1)]]" {
		t.Errorf("DivScalar = %s", dv.ToS())
	}
	div, _ := a.Div(Identity(2))
	if div.ToS() != "Matrix[[(1/1), (2/1)], [(3/1), (4/1)]]" {
		t.Errorf("Div = %s", div.ToS())
	}
	if a.Transpose().ToS() != "Matrix[[1, 3], [2, 4]]" {
		t.Errorf("Transpose = %s", a.Transpose().ToS())
	}
	tr, _ := a.Trace()
	if tr.String() != "5" {
		t.Errorf("Trace = %s", tr)
	}
}

func TestArithmeticErrors(t *testing.T) {
	a := mat(t, [][]any{{1, 2}})
	b := mat(t, [][]any{{1, 2, 3}})
	if _, err := a.Add(b); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Add mismatch = %v", err)
	}
	if _, err := a.Sub(b); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Sub mismatch = %v", err)
	}
	if _, err := a.Mul(a); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Mul mismatch = %v", err)
	}
	if _, err := a.MulVector(vec(t, 1, 2, 3, 4)); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("MulVector mismatch = %v", err)
	}
	if _, err := a.MulScalar("x"); err == nil {
		t.Error("MulScalar bad")
	}
	if _, err := a.DivScalar("x"); err == nil {
		t.Error("DivScalar bad")
	}
	ns := mat(t, [][]any{{1, 2, 3}})
	if _, err := ns.Trace(); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Trace nonsquare = %v", err)
	}
	if _, err := ns.Div(Identity(2)); err == nil {
		t.Error("Div of nonsquare divisor")
	}
}

func TestPow(t *testing.T) {
	a := mat(t, [][]any{{1, 2}, {3, 4}})
	p2, _ := a.Pow(2)
	if p2.ToS() != "Matrix[[7, 10], [15, 22]]" {
		t.Errorf("Pow2 = %s", p2.ToS())
	}
	p0, _ := a.Pow(0)
	if p0.ToS() != "Matrix[[1, 0], [0, 1]]" {
		t.Errorf("Pow0 = %s", p0.ToS())
	}
	p3, _ := a.Pow(3)
	if p3.ToS() != "Matrix[[37, 54], [81, 118]]" {
		t.Errorf("Pow3 = %s", p3.ToS())
	}
	pn1, _ := a.Pow(-1)
	if pn1.ToS() != "Matrix[[(-2/1), (1/1)], [(3/2), (-1/2)]]" {
		t.Errorf("Pow-1 = %s", pn1.ToS())
	}
	if _, err := mat(t, [][]any{{1, 2, 3}}).Pow(2); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Pow nonsquare = %v", err)
	}
	if _, err := mat(t, [][]any{{1, 1}, {1, 1}}).Pow(-1); !errors.Is(err, ErrNotRegular) {
		t.Errorf("Pow-1 singular = %v", err)
	}
}

func TestDeterminant(t *testing.T) {
	cases := []struct {
		rows [][]any
		want string
	}{
		{[][]any{{1, 2}, {3, 4}}, "-2"},
		{[][]any{{5}}, "5"},
		{[][]any{{2, 0, 0}, {0, 3, 0}, {0, 0, 4}}, "24"},
		{[][]any{{1, 2, 3}, {4, 5, 6}, {7, 8, 10}}, "-3"},
		{[][]any{{1.0, 2.0}, {3.0, 4.0}}, "-2.0"},
	}
	for _, c := range cases {
		d, err := mat(t, c.rows).Determinant()
		if err != nil || d.String() != c.want {
			t.Errorf("det(%v) = %s, %v; want %s", c.rows, d, err, c.want)
		}
	}
	if d, _ := Zero(0, 0).Determinant(); d.String() != "1" {
		t.Errorf("det empty = %s", d)
	}
	if _, err := mat(t, [][]any{{1, 2, 3}}).Determinant(); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("det nonsquare = %v", err)
	}
}

func TestInverse(t *testing.T) {
	cases := []struct {
		rows [][]any
		want string
	}{
		{[][]any{{1, 2}, {3, 4}}, "Matrix[[(-2/1), (1/1)], [(3/2), (-1/2)]]"},
		{[][]any{{2, 0}, {0, 2}}, "Matrix[[(1/2), (0/1)], [(0/1), (1/2)]]"},
		{[][]any{{4, 7}, {2, 6}}, "Matrix[[(3/5), (-7/10)], [(-1/5), (2/5)]]"},
		{[][]any{{1, 2, 3}, {0, 1, 4}, {5, 6, 0}}, "Matrix[[(-24/1), (18/1), (5/1)], [(20/1), (-15/1), (-4/1)], [(-5/1), (4/1), (1/1)]]"},
		{[][]any{{1.0, 2.0}, {3.0, 4.0}}, "Matrix[[-1.9999999999999998, 0.9999999999999998], [1.4999999999999998, -0.49999999999999994]]"},
	}
	for _, c := range cases {
		inv, err := mat(t, c.rows).Inverse()
		if err != nil || inv.ToS() != c.want {
			t.Errorf("inv(%v) = %s, %v; want %s", c.rows, inv.ToS(), err, c.want)
		}
	}
	if _, err := mat(t, [][]any{{1, 2, 3}}).Inverse(); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("inv nonsquare = %v", err)
	}
	if _, err := mat(t, [][]any{{1, 2}, {2, 4}}).Inverse(); !errors.Is(err, ErrNotRegular) {
		t.Errorf("inv singular = %v", err)
	}
}

func TestRank(t *testing.T) {
	cases := []struct {
		rows [][]any
		want int
	}{
		{[][]any{{1, 2}, {3, 4}}, 2},
		{[][]any{{1, 2, 3}, {2, 4, 6}, {1, 0, 1}}, 2},
		{[][]any{{1, 2, 3}, {2, 4, 6}}, 1},
		{[][]any{{0, 0}, {0, 0}}, 0},
		{[][]any{{1, 0, 0}, {0, 1, 0}}, 2},
	}
	for _, c := range cases {
		if got := mat(t, c.rows).Rank(); got != c.want {
			t.Errorf("rank(%v) = %d; want %d", c.rows, got, c.want)
		}
	}
	if got := Zero(0, 0).Rank(); got != 0 {
		t.Errorf("rank empty = %d", got)
	}
}

func TestPredicates(t *testing.T) {
	if !mat(t, [][]any{{1, 2}, {3, 4}}).Square() {
		t.Error("Square")
	}
	if mat(t, [][]any{{1, 2, 3}}).Square() {
		t.Error("non-Square")
	}
	if !Zero(2, 2).IsZero() {
		t.Error("IsZero")
	}
	if mat(t, [][]any{{1, 0}}).IsZero() {
		t.Error("not IsZero")
	}
	if !mat(t, [][]any{{1, 0}, {0, 2}}).IsDiagonal() {
		t.Error("IsDiagonal")
	}
	if mat(t, [][]any{{1, 1}, {0, 2}}).IsDiagonal() {
		t.Error("not diagonal off-entry")
	}
	if mat(t, [][]any{{1, 2, 3}}).IsDiagonal() {
		t.Error("diagonal nonsquare")
	}
	if !mat(t, [][]any{{1, 2}, {2, 1}}).Symmetric() {
		t.Error("Symmetric")
	}
	if mat(t, [][]any{{1, 2}, {3, 1}}).Symmetric() {
		t.Error("not symmetric")
	}
	if mat(t, [][]any{{1, 2, 3}}).Symmetric() {
		t.Error("symmetric nonsquare")
	}
	if !mat(t, [][]any{{1, 0}, {3, 4}}).LowerTriangular() {
		t.Error("LowerTriangular")
	}
	if mat(t, [][]any{{1, 2}, {3, 4}}).LowerTriangular() {
		t.Error("not lower")
	}
	if !mat(t, [][]any{{1, 2}, {0, 4}}).UpperTriangular() {
		t.Error("UpperTriangular")
	}
	if mat(t, [][]any{{1, 2}, {3, 4}}).UpperTriangular() {
		t.Error("not upper")
	}
}

func TestRegularSingularOrthogonal(t *testing.T) {
	reg, err := mat(t, [][]any{{1, 2}, {3, 4}}).Regular()
	if err != nil || !reg {
		t.Errorf("Regular = %v, %v", reg, err)
	}
	sing, err := mat(t, [][]any{{1, 2}, {2, 4}}).Singular()
	if err != nil || !sing {
		t.Errorf("Singular = %v, %v", sing, err)
	}
	if _, err := mat(t, [][]any{{1, 2, 3}}).Regular(); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Regular nonsquare = %v", err)
	}
	if _, err := mat(t, [][]any{{1, 2, 3}}).Singular(); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Singular nonsquare = %v", err)
	}
	orth, err := mat(t, [][]any{{0, 1}, {1, 0}}).Orthogonal()
	if err != nil || !orth {
		t.Errorf("Orthogonal = %v, %v", orth, err)
	}
	notOrth, err := mat(t, [][]any{{1, 2}, {3, 4}}).Orthogonal()
	if err != nil || notOrth {
		t.Errorf("not Orthogonal = %v, %v", notOrth, err)
	}
	if o, err := mat(t, [][]any{{1, 2, 3}}).Orthogonal(); err != nil || o {
		t.Errorf("Orthogonal nonsquare = %v, %v", o, err)
	}
}

func TestEqlHashRound(t *testing.T) {
	a := mat(t, [][]any{{1, 2}, {3, 4}})
	if !a.Eql(mat(t, [][]any{{1, 2}, {3, 4}})) {
		t.Error("Eql equal")
	}
	if a.Eql(mat(t, [][]any{{1, 2}})) {
		t.Error("Eql diff shape")
	}
	if a.Eql(mat(t, [][]any{{1, 2}, {3, 5}})) {
		t.Error("Eql diff entry")
	}
	// cross-kind equality: 2 == (2/1) == 2.0
	mInt := mat(t, [][]any{{2}})
	mRat := newFromNums(1, 1, []Num{NewRat(2, 1)})
	mFlt := mat(t, [][]any{{2.0}})
	if !mInt.Eql(mRat) || !mInt.Eql(mFlt) {
		t.Error("cross-kind Eql")
	}
	if a.Hash() != mat(t, [][]any{{1, 2}, {3, 4}}).Hash() {
		t.Error("Hash equal matrices differ")
	}
	if a.Hash() == mat(t, [][]any{{9, 9}, {9, 9}}).Hash() {
		t.Error("Hash collision expected distinct")
	}
	fm := mat(t, [][]any{{1.234, 2.567}})
	if fm.RoundEntries(1).ToS() != "Matrix[[1.2, 2.6]]" {
		t.Errorf("RoundEntries = %s", fm.RoundEntries(1).ToS())
	}
	// rounding leaves Integer/Rational entries unchanged
	if a.RoundEntries(2).ToS() != "Matrix[[1, 2], [3, 4]]" {
		t.Errorf("RoundEntries int = %s", a.RoundEntries(2).ToS())
	}
}

func TestEmptyToS(t *testing.T) {
	if got := Zero(2, 0).ToS(); got != "Matrix.empty(2, 0)" {
		t.Errorf("empty ToS = %s", got)
	}
	if got := Zero(2, 0).Inspect(); got != "Matrix.empty(2, 0)" {
		t.Errorf("empty Inspect = %s", got)
	}
	if got := mat(t, [][]any{{1}}).Inspect(); got != "Matrix[[1]]" {
		t.Errorf("Inspect = %s", got)
	}
}
