// Copyright (c) the go-ruby-matrix/matrix authors
//
// SPDX-License-Identifier: BSD-3-Clause

package matrix

import (
	"math"
	"math/big"
	"testing"
)

func TestNumPromotion(t *testing.T) {
	// Integer op Integer stays Integer.
	if got := NewInt(2).Add(NewInt(3)); got.kind != kindInt || got.String() != "5" {
		t.Errorf("int+int = %s (kind %d)", got, got.kind)
	}
	if got := NewInt(7).Sub(NewInt(2)); got.String() != "5" {
		t.Errorf("int-int = %s", got)
	}
	if got := NewInt(2).Mul(NewInt(3)); got.String() != "6" {
		t.Errorf("int*int = %s", got)
	}
	// Rational stays Rational even when whole.
	if got := NewRat(2, 1).Add(NewInt(3)); got.kind != kindRat || got.String() != "(5/1)" {
		t.Errorf("rat+int = %s", got)
	}
	if got := NewRat(3, 4).Sub(NewRat(1, 4)); got.String() != "(1/2)" {
		t.Errorf("rat-rat = %s", got)
	}
	if got := NewRat(2, 3).Mul(NewRat(3, 2)); got.String() != "(1/1)" {
		t.Errorf("rat*rat = %s", got)
	}
	// Float dominates.
	if got := NewRat(1, 2).Add(NewFloat(0.5)); got.kind != kindFlt || got.String() != "1.0" {
		t.Errorf("rat+float = %s", got)
	}
	if got := NewFloat(1.0).Sub(NewInt(2)); got.String() != "-1.0" {
		t.Errorf("float-int = %s", got)
	}
	if got := NewFloat(2.0).Mul(NewInt(3)); got.String() != "6.0" {
		t.Errorf("float*int = %s", got)
	}
	// Quo: int/int -> Rational; float involvement -> Float.
	if got := NewInt(1).Quo(NewInt(2)); got.kind != kindRat || got.String() != "(1/2)" {
		t.Errorf("int.quo = %s", got)
	}
	if got := NewFloat(1.0).Quo(NewInt(2)); got.String() != "0.5" {
		t.Errorf("float.quo = %s", got)
	}
}

func TestNumNegZeroEql(t *testing.T) {
	if NewInt(-3).Neg().String() != "3" {
		t.Error("int neg")
	}
	if NewRat(1, 2).Neg().String() != "(-1/2)" {
		t.Error("rat neg")
	}
	if NewFloat(2.5).Neg().String() != "-2.5" {
		t.Error("float neg")
	}
	if !NewInt(0).IsZero() || !NewRat(0, 1).IsZero() || !NewFloat(0).IsZero() {
		t.Error("IsZero kinds")
	}
	if NewInt(1).IsZero() || NewRat(1, 2).IsZero() || NewFloat(0.1).IsZero() {
		t.Error("not zero kinds")
	}
	// cross-kind equality
	if !NewInt(2).Eql(NewRat(2, 1)) || !NewInt(2).Eql(NewFloat(2.0)) {
		t.Error("cross-kind eql")
	}
	if NewInt(2).Eql(NewInt(3)) {
		t.Error("int eql diff")
	}
}

func TestNumBigAndSqrtRound(t *testing.T) {
	bi, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	n := NewBigInt(bi)
	if n.String() != "123456789012345678901234567890" {
		t.Errorf("bigint = %s", n)
	}
	r := NewBigRat(big.NewRat(3, 6)) // reduces to 1/2
	if r.String() != "(1/2)" {
		t.Errorf("bigrat = %s", r)
	}
	if NewInt(9).Sqrt().String() != "3.0" {
		t.Errorf("sqrt = %s", NewInt(9).Sqrt())
	}
	// Round leaves non-Float untouched, rounds Float.
	if NewInt(5).Round(2).String() != "5" {
		t.Error("round int")
	}
	if NewRat(1, 2).Round(2).String() != "(1/2)" {
		t.Error("round rat")
	}
	if NewFloat(1.2345).Round(2).String() != "1.23" {
		t.Errorf("round float = %s", NewFloat(1.2345).Round(2))
	}
}

func TestFloatFormatting(t *testing.T) {
	cases := []struct {
		f    float64
		want string
	}{
		{2.0, "2.0"},
		{0.6, "0.6"},
		{0.8, "0.8"},
		{5.0, "5.0"},
		{-1.9999999999999998, "-1.9999999999999998"},
		{0.9999999999999998, "0.9999999999999998"},
		{1.4999999999999998, "1.4999999999999998"},
		{-0.49999999999999994, "-0.49999999999999994"},
		{100000000000000.0, "100000000000000.0"},
		{123456789012345.6, "123456789012345.6"},
		{1e15, "1.0e+15"},
		{1e16, "1.0e+16"},
		{1e17, "1.0e+17"},
		{1e20, "1.0e+20"},
		{1e21, "1.0e+21"},
		{1.2345678901234568e+16, "1.2345678901234568e+16"},
		{9.999999999999998e+15, "9.999999999999998e+15"},
		{0.0001, "0.0001"},
		{1e-5, "1.0e-05"},
		{1.5e-10, "1.5e-10"},
		{0.0, "0.0"},
		{math.Copysign(0, -1), "-0.0"},
		{math.Inf(1), "Infinity"},
		{math.Inf(-1), "-Infinity"},
		{math.NaN(), "NaN"},
		{1234567.89, "1234567.89"},
	}
	for _, c := range cases {
		if got := formatRubyFloat(c.f); got != c.want {
			t.Errorf("formatRubyFloat(%v) = %q; want %q", c.f, got, c.want)
		}
	}
}

func TestItoa(t *testing.T) {
	cases := map[int]string{0: "0", 5: "5", 42: "42", -7: "-7", -100: "-100"}
	for in, want := range cases {
		if got := itoa(in); got != want {
			t.Errorf("itoa(%d) = %q; want %q", in, got, want)
		}
	}
}

// TestNumDiv exercises Num.Div, which follows Ruby's `/` operator per operand
// kind (Integer/Integer floors, any Rational stays Rational, any Float yields a
// Float) — the semantics Matrix#/ with a scalar must match against MRI 4.0.5.
func TestNumDiv(t *testing.T) {
	cases := []struct {
		a, b Num
		want string
	}{
		// Integer / Integer: floor division (toward negative infinity).
		{NewInt(3), NewInt(2), "1"},
		{NewInt(7), NewInt(2), "3"},
		{NewInt(-3), NewInt(2), "-2"},
		{NewInt(-7), NewInt(2), "-4"},
		{NewInt(7), NewInt(-2), "-4"},
		{NewInt(-7), NewInt(-2), "3"},
		{NewInt(6), NewInt(3), "2"}, // exact: no flooring correction
		{NewInt(-6), NewInt(3), "-2"},
		// Float involvement: Float result.
		{NewInt(3), NewFloat(2.0), "1.5"},
		{NewFloat(3.0), NewInt(2), "1.5"},
		// Rational (no Float): Rational result.
		{NewInt(3), NewRat(2, 1), "(3/2)"},
		{NewRat(3, 1), NewInt(2), "(3/2)"},
		{NewRat(1, 2), NewRat(1, 4), "(2/1)"},
	}
	for _, c := range cases {
		if got := c.a.Div(c.b).String(); got != c.want {
			t.Errorf("%v.Div(%v) = %s; want %s", c.a, c.b, got, c.want)
		}
	}
}

// TestNumRoundKinds covers Num.Round across kinds and the ndigits sign rules:
// ndigits <= 0 (including the no-arg form, called as Round(0)) yields Integer;
// ndigits >= 1 keeps the operand kind. All halves round away from zero, matching
// MRI 4.0.5.
func TestNumRoundKinds(t *testing.T) {
	cases := []struct {
		n       Num
		ndigits int
		want    string
		kind    numKind
	}{
		// no-arg / round(0): Integer result for every kind.
		{NewFloat(1.4), 0, "1", kindInt},
		{NewFloat(2.6), 0, "3", kindInt},
		{NewFloat(-2.5), 0, "-3", kindInt}, // half away from zero
		{NewFloat(2.5), 0, "3", kindInt},
		{NewInt(5), 0, "5", kindInt},
		{NewRat(3, 2), 0, "2", kindInt},
		{NewRat(7, 2), 0, "4", kindInt},
		{NewRat(-3, 2), 0, "-2", kindInt},
		// ndigits < 0: Integer, rounding at the 10**(-n) place.
		{NewFloat(14.5), -1, "10", kindInt},
		{NewFloat(25.5), -1, "30", kindInt},
		{NewInt(15), -1, "20", kindInt},
		{NewInt(14), -1, "10", kindInt},
		{NewRat(255, 10), -1, "30", kindInt},
		// ndigits >= 1: kind preserved.
		{NewInt(5), 2, "5", kindInt},
		{NewRat(1, 2), 2, "(1/2)", kindRat},
		{NewRat(7, 3), 1, "(23/10)", kindRat},
		{NewFloat(1.2345), 2, "1.23", kindFlt},
		{NewFloat(2.567), 1, "2.6", kindFlt},
	}
	for _, c := range cases {
		got := c.n.Round(c.ndigits)
		if got.String() != c.want || got.kind != c.kind {
			t.Errorf("%v.Round(%d) = %s (kind %d); want %s (kind %d)",
				c.n, c.ndigits, got.String(), got.kind, c.want, c.kind)
		}
	}
}

func TestDivSingularDivisor(t *testing.T) {
	a, _ := New([][]any{{1, 2}, {3, 4}})
	sing, _ := New([][]any{{1, 1}, {1, 1}})
	if _, err := a.Div(sing); err == nil {
		t.Error("Div by singular: want error")
	}
}
