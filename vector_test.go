// Copyright (c) the go-ruby-matrix/matrix authors
//
// SPDX-License-Identifier: BSD-3-Clause

package matrix

import (
	"errors"
	"math"
	"testing"
)

func TestVectorBasics(t *testing.T) {
	v := vec(t, 10, 20, 30)
	if v.Size() != 3 {
		t.Errorf("Size = %d", v.Size())
	}
	if n, ok := v.At(1); !ok || n.String() != "20" {
		t.Errorf("At(1) = %s, %v", n, ok)
	}
	if _, ok := v.At(9); ok {
		t.Error("At oob")
	}
	if _, ok := v.At(-1); ok {
		t.Error("At negative")
	}
	els := v.Elements()
	if len(els) != 3 || els[2].String() != "30" {
		t.Errorf("Elements = %v", els)
	}
	var sum int64
	v.Each(func(n Num) { sum += n.i.Int64() })
	if sum != 60 {
		t.Errorf("Each sum = %d", sum)
	}
	if v.ToS() != "Vector[10, 20, 30]" || v.Inspect() != "Vector[10, 20, 30]" {
		t.Errorf("ToS/Inspect = %s / %s", v.ToS(), v.Inspect())
	}
}

func TestVectorNewError(t *testing.T) {
	if _, err := NewVector([]any{"x"}); err == nil {
		t.Error("NewVector bad: want error")
	}
}

func TestVectorMap(t *testing.T) {
	v := vec(t, 1, 2, 3)
	m, err := v.Map(func(n Num) any { return n.Mul(NewInt(10)) })
	if err != nil || m.ToS() != "Vector[10, 20, 30]" {
		t.Errorf("Map = %s, %v", m.ToS(), err)
	}
	if _, err := v.Map(func(n Num) any { return "x" }); err == nil {
		t.Error("Map bad result: want error")
	}
}

func TestVectorArithmetic(t *testing.T) {
	a := vec(t, 1, 2)
	b := vec(t, 3, 4)
	add, _ := a.Add(b)
	if add.ToS() != "Vector[4, 6]" {
		t.Errorf("Add = %s", add.ToS())
	}
	sub, _ := a.Sub(b)
	if sub.ToS() != "Vector[-2, -2]" {
		t.Errorf("Sub = %s", sub.ToS())
	}
	mul, _ := a.Mul(3)
	if mul.ToS() != "Vector[3, 6]" {
		t.Errorf("Mul = %s", mul.ToS())
	}
	if _, err := a.Add(vec(t, 1, 2, 3)); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Add mismatch = %v", err)
	}
	if _, err := a.Sub(vec(t, 1, 2, 3)); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Sub mismatch = %v", err)
	}
	if _, err := a.Mul("x"); err == nil {
		t.Error("Mul bad scalar")
	}
}

func TestVectorProducts(t *testing.T) {
	dot, err := vec(t, 1, 2, 3).InnerProduct(vec(t, 4, 5, 6))
	if err != nil || dot.String() != "32" {
		t.Errorf("InnerProduct = %s, %v", dot, err)
	}
	if _, err := vec(t, 1).InnerProduct(vec(t, 1, 2)); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("InnerProduct mismatch = %v", err)
	}
	cp, err := vec(t, 1, 0, 0).CrossProduct(vec(t, 0, 1, 0))
	if err != nil || cp.ToS() != "Vector[0, 0, 1]" {
		t.Errorf("CrossProduct = %s, %v", cp.ToS(), err)
	}
	if _, err := vec(t, 1, 2).CrossProduct(vec(t, 1, 2)); !errors.Is(err, ErrOperationNotDefined) {
		t.Errorf("CrossProduct 2D = %v", err)
	}
}

func TestVectorMagnitudeNormalize(t *testing.T) {
	mag := vec(t, 3, 4).Magnitude()
	if mag.String() != "5.0" {
		t.Errorf("Magnitude = %s", mag)
	}
	norm, err := vec(t, 3, 4).Normalize()
	if err != nil || norm.ToS() != "Vector[0.6, 0.8]" {
		t.Errorf("Normalize = %s, %v", norm.ToS(), err)
	}
	if _, err := vec(t, 0, 0).Normalize(); !errors.Is(err, ErrOperationNotDefined) {
		t.Errorf("Normalize zero = %v", err)
	}
}

func TestVectorAngle(t *testing.T) {
	ang, err := vec(t, 1, 0).Angle(vec(t, 0, 1))
	if err != nil || math.Abs(ang.asFloat()-math.Pi/2) > 1e-12 {
		t.Errorf("Angle = %v, %v", ang, err)
	}
	// parallel -> 0
	ang0, _ := vec(t, 1, 0).Angle(vec(t, 2, 0))
	if math.Abs(ang0.asFloat()) > 1e-12 {
		t.Errorf("Angle parallel = %v", ang0)
	}
	// anti-parallel: float rounding makes cos slightly < -1, exercising the lower
	// clamp; the angle is pi.
	angPi, _ := vec(t, 1, 1, 1).Angle(vec(t, -1, -1, -1))
	if math.Abs(angPi.asFloat()-math.Pi) > 1e-12 {
		t.Errorf("Angle anti = %v", angPi)
	}
	// angle of a vector with itself: float rounding makes cos slightly > 1, so
	// this exercises the upper clamp; MRI returns exactly 0.
	angSelf, _ := vec(t, 1, 1, 1).Angle(vec(t, 1, 1, 1))
	if angSelf.asFloat() != 0 {
		t.Errorf("Angle self = %v; want 0", angSelf)
	}
	if _, err := vec(t, 1, 0).Angle(vec(t, 1, 2, 3)); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Angle mismatch = %v", err)
	}
	if _, err := vec(t, 0, 0).Angle(vec(t, 1, 1)); !errors.Is(err, ErrOperationNotDefined) {
		t.Errorf("Angle zero = %v", err)
	}
}

func TestVectorEql(t *testing.T) {
	if !vec(t, 1, 2).Eql(vec(t, 1, 2)) {
		t.Error("Eql equal")
	}
	if vec(t, 1, 2).Eql(vec(t, 1, 2, 3)) {
		t.Error("Eql diff size")
	}
	if vec(t, 1, 2).Eql(vec(t, 1, 3)) {
		t.Error("Eql diff entry")
	}
}

func TestIndependent(t *testing.T) {
	ind, err := Independent(vec(t, 1, 0), vec(t, 0, 1))
	if err != nil || !ind {
		t.Errorf("Independent = %v, %v", ind, err)
	}
	dep, err := Independent(vec(t, 1, 2), vec(t, 2, 4))
	if err != nil || dep {
		t.Errorf("dependent = %v, %v", dep, err)
	}
	// more vectors than dimensions -> dependent
	over, err := Independent(vec(t, 1, 0), vec(t, 0, 1), vec(t, 1, 1))
	if err != nil || over {
		t.Errorf("over-determined = %v, %v", over, err)
	}
	// differing sizes -> error
	if _, err := Independent(vec(t, 1, 0), vec(t, 0, 1, 2)); !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Independent mismatch = %v", err)
	}
	// empty set -> trivially independent
	if ok, err := Independent(); err != nil || !ok {
		t.Errorf("Independent() = %v, %v", ok, err)
	}
}
