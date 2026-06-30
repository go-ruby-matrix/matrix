// Copyright (c) the go-ruby-matrix/matrix authors
//
// SPDX-License-Identifier: BSD-3-Clause

package matrix

import (
	"fmt"
	"math"
	"math/big"
)

// Num is one entry of a Matrix or Vector. It models the slice of Ruby's numeric
// tower that the `matrix` stdlib actually depends on: Integer, Rational and
// Float. Keeping the three kinds distinct is what lets this package reproduce
// MRI's exact-arithmetic output byte for byte — `Matrix[[1,2],[3,4]].inverse`
// must print exact Rationals like `(-2/1)`, and `det` of an Integer matrix must
// stay an Integer, never silently degrade to Float.
//
// The kinds promote following Ruby's coercion rules:
//
//   - Integer op Integer            → Integer   (except Quo, see below)
//   - any Rational, no Float        → Rational  (stays Rational even when whole)
//   - any Float                     → Float
//
// Quo is exact "/": Integer.quo(Integer) yields a Rational, matching how MRI's
// Matrix#inverse and Vector#normalize divide. Plain Ruby Integer "/" floors, but
// the matrix library never uses it, so Num offers only the exact Quo.
type Num struct {
	kind numKind
	i    *big.Int // kind == kindInt
	r    *big.Rat // kind == kindRat
	f    float64  // kind == kindFlt
}

type numKind uint8

const (
	kindInt numKind = iota
	kindRat
	kindFlt
)

// NewInt returns an Integer Num.
func NewInt(v int64) Num { return Num{kind: kindInt, i: big.NewInt(v)} }

// NewBigInt returns an Integer Num from a *big.Int (the value is copied).
func NewBigInt(v *big.Int) Num { return Num{kind: kindInt, i: new(big.Int).Set(v)} }

// NewRat returns a Rational Num n/d (reduced; d must be non-zero).
func NewRat(n, d int64) Num { return Num{kind: kindRat, r: big.NewRat(n, d)} }

// NewBigRat returns a Rational Num from a *big.Rat (the value is copied).
func NewBigRat(v *big.Rat) Num { return Num{kind: kindRat, r: new(big.Rat).Set(v)} }

// NewFloat returns a Float Num.
func NewFloat(v float64) Num { return Num{kind: kindFlt, f: v} }

// numFromAny converts a Go value supplied by a caller (or rbgo) into a Num. It
// accepts the integer kinds, *big.Int, *big.Rat, float kinds, and a Num itself.
func numFromAny(v any) (Num, error) {
	switch x := v.(type) {
	case Num:
		return x, nil
	case int:
		return NewInt(int64(x)), nil
	case int8:
		return NewInt(int64(x)), nil
	case int16:
		return NewInt(int64(x)), nil
	case int32:
		return NewInt(int64(x)), nil
	case int64:
		return NewInt(x), nil
	case uint:
		return NewBigInt(new(big.Int).SetUint64(uint64(x))), nil
	case uint64:
		return NewBigInt(new(big.Int).SetUint64(x)), nil
	case *big.Int:
		return NewBigInt(x), nil
	case *big.Rat:
		return NewBigRat(x), nil
	case float32:
		return NewFloat(float64(x)), nil
	case float64:
		return NewFloat(x), nil
	default:
		return Num{}, fmt.Errorf("matrix: cannot use %T (%v) as a numeric entry", v, v)
	}
}

// asRat returns the value of an Integer or Rational Num as a *big.Rat. It is
// only ever called after coerce() has ruled out Float, so an Integer Num is the
// sole non-Rational case.
func (n Num) asRat() *big.Rat {
	if n.kind == kindRat {
		return new(big.Rat).Set(n.r)
	}
	return new(big.Rat).SetInt(n.i)
}

// asFloat returns the value of any Num as a float64.
func (n Num) asFloat() float64 {
	switch n.kind {
	case kindInt:
		f := new(big.Float).SetInt(n.i)
		v, _ := f.Float64()
		return v
	case kindRat:
		v, _ := n.r.Float64()
		return v
	default:
		return n.f
	}
}

// ratToNum wraps a reduced *big.Rat back into a Num, keeping it Rational. Unlike
// Ruby's Rational#to_i, MRI's Matrix keeps whole Rationals as Rationals (the
// inverse of [[2,0],[0,2]] prints `(1/2)` and `(0/1)`, not `0`), so this never
// collapses to Integer.
func ratToNum(r *big.Rat) Num { return Num{kind: kindRat, r: r} }

// binop applies one of the basic arithmetic operations following Ruby's coercion
// rules. exact selects whether Int/Int division is allowed to yield a Rational.
func (a Num) coerce(b Num) numKind {
	if a.kind == kindFlt || b.kind == kindFlt {
		return kindFlt
	}
	if a.kind == kindRat || b.kind == kindRat {
		return kindRat
	}
	return kindInt
}

// Add returns a+b.
func (a Num) Add(b Num) Num {
	switch a.coerce(b) {
	case kindFlt:
		return NewFloat(a.asFloat() + b.asFloat())
	case kindRat:
		return ratToNum(new(big.Rat).Add(a.asRat(), b.asRat()))
	default:
		return Num{kind: kindInt, i: new(big.Int).Add(a.i, b.i)}
	}
}

// Sub returns a-b.
func (a Num) Sub(b Num) Num {
	switch a.coerce(b) {
	case kindFlt:
		return NewFloat(a.asFloat() - b.asFloat())
	case kindRat:
		return ratToNum(new(big.Rat).Sub(a.asRat(), b.asRat()))
	default:
		return Num{kind: kindInt, i: new(big.Int).Sub(a.i, b.i)}
	}
}

// Mul returns a*b.
func (a Num) Mul(b Num) Num {
	switch a.coerce(b) {
	case kindFlt:
		return NewFloat(a.asFloat() * b.asFloat())
	case kindRat:
		return ratToNum(new(big.Rat).Mul(a.asRat(), b.asRat()))
	default:
		return Num{kind: kindInt, i: new(big.Int).Mul(a.i, b.i)}
	}
}

// Quo returns the exact quotient a/b. Integer/Integer yields a Rational (Ruby's
// Integer#quo), mirroring how Matrix#inverse divides; Float involvement yields a
// Float. b must be non-zero (callers guard against a zero pivot/divisor).
func (a Num) Quo(b Num) Num {
	if a.kind == kindFlt || b.kind == kindFlt {
		return NewFloat(a.asFloat() / b.asFloat())
	}
	return ratToNum(new(big.Rat).Quo(a.asRat(), b.asRat()))
}

// Div returns a/b following Ruby's `/` operator per operand kind, as used by
// `Matrix#/` with a scalar. Unlike the exact Quo, this reproduces Ruby's
// type-directed division:
//
//   - Integer / Integer → Integer, floor division (so 3/2 == 1, -3/2 == -2);
//   - Integer / Rational, Rational / anything (non-Float) → Rational;
//   - any Float involvement → Float.
//
// b must be non-zero (callers guard against a zero divisor).
func (a Num) Div(b Num) Num {
	switch {
	case a.kind == kindFlt || b.kind == kindFlt:
		return NewFloat(a.asFloat() / b.asFloat())
	case a.kind == kindRat || b.kind == kindRat:
		return ratToNum(new(big.Rat).Quo(a.asRat(), b.asRat()))
	default:
		// Integer / Integer floors toward negative infinity, like Ruby's
		// Integer#/. big.Int.QuoRem truncates toward zero, so when the
		// truncated remainder is non-zero and the operands have opposite signs
		// the quotient is one too high and must be decremented to floor it.
		q := new(big.Int)
		r := new(big.Int)
		q.QuoRem(a.i, b.i, r)
		if r.Sign() != 0 && (a.i.Sign() < 0) != (b.i.Sign() < 0) {
			q.Sub(q, big.NewInt(1))
		}
		return Num{kind: kindInt, i: q}
	}
}

// Neg returns -a.
func (a Num) Neg() Num {
	switch a.kind {
	case kindFlt:
		return NewFloat(-a.f)
	case kindRat:
		return ratToNum(new(big.Rat).Neg(a.r))
	default:
		return Num{kind: kindInt, i: new(big.Int).Neg(a.i)}
	}
}

// IsZero reports whether a == 0.
func (a Num) IsZero() bool {
	switch a.kind {
	case kindFlt:
		return a.f == 0
	case kindRat:
		return a.r.Sign() == 0
	default:
		return a.i.Sign() == 0
	}
}

// Eql reports numeric equality across kinds, matching Ruby's == (Rational(2,1)
// equals the Integer 2, and 2.0 equals 2).
func (a Num) Eql(b Num) bool {
	if a.kind == kindFlt || b.kind == kindFlt {
		return a.asFloat() == b.asFloat()
	}
	return a.asRat().Cmp(b.asRat()) == 0
}

// Sqrt returns the Float square root of a (always a Float, like Ruby's
// Integer#** with 0.5 / Math.sqrt used by Vector#magnitude).
func (a Num) Sqrt() Num { return NewFloat(math.Sqrt(a.asFloat())) }

// Round returns a rounded to ndigits decimal places, following Ruby's
// Integer#round / Rational#round / Float#round, which `Matrix#round` maps over
// the entries. Two rules drive the result kind, exactly as in MRI:
//
//   - ndigits <= 0 always yields an Integer (this includes the no-arg form,
//     called here as Round(0)), rounding half away from zero at the
//     10**(-ndigits) place;
//   - ndigits >= 1 keeps the operand's kind: an Integer stays itself, a
//     Rational stays a (rounded) Rational, a Float stays a (rounded) Float.
//
// All kinds round half away from zero, matching Ruby's default rounding mode.
func (a Num) Round(ndigits int) Num {
	if ndigits >= 1 {
		switch a.kind {
		case kindInt:
			return a
		case kindRat:
			return ratToNum(roundRat(a.r, ndigits))
		default:
			p := math.Pow(10, float64(ndigits))
			return NewFloat(roundHalfAway(a.f*p) / p)
		}
	}
	// ndigits <= 0: Integer result.
	if a.kind == kindFlt {
		p := math.Pow(10, float64(-ndigits))
		return NewBigInt(floatToBigInt(roundHalfAway(a.f/p) * p))
	}
	// Integer or Rational: round exactly to the 10**(-ndigits) place.
	return NewBigInt(roundRat(a.asRat(), ndigits).Num())
}

// roundRat rounds r to ndigits decimal places (half away from zero) and returns
// the result as an exact *big.Rat. For ndigits >= 0 the result is r scaled,
// rounded to an integer, then unscaled; for ndigits <= 0 the rounded integer is
// scaled back up, leaving a whole value whose Num() is the Integer answer.
func roundRat(r *big.Rat, ndigits int) *big.Rat {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(abs(ndigits))), nil)
	scaleRat := new(big.Rat).SetInt(scale)
	scaled := new(big.Rat)
	if ndigits >= 0 {
		scaled.Mul(r, scaleRat)
	} else {
		scaled.Quo(r, scaleRat)
	}
	rounded := new(big.Rat).SetInt(ratRoundHalfAway(scaled))
	if ndigits >= 0 {
		return rounded.Quo(rounded, scaleRat)
	}
	return rounded.Mul(rounded, scaleRat)
}

// ratRoundHalfAway rounds a *big.Rat to the nearest integer, with halves going
// away from zero (matching Ruby's default).
func ratRoundHalfAway(r *big.Rat) *big.Int {
	num := r.Num()
	den := r.Denom()
	q := new(big.Int)
	rem := new(big.Int)
	q.QuoRem(num, den, rem) // truncated toward zero
	if rem.Sign() == 0 {
		return q
	}
	// Compare 2*|rem| with den to decide whether to round away from zero.
	twice := new(big.Int).Abs(rem)
	twice.Lsh(twice, 1)
	if twice.Cmp(den) >= 0 {
		if num.Sign() < 0 {
			q.Sub(q, big.NewInt(1))
		} else {
			q.Add(q, big.NewInt(1))
		}
	}
	return q
}

// roundHalfAway rounds a float64 to the nearest integer, halves away from zero.
// math.Round already implements this rule.
func roundHalfAway(f float64) float64 { return math.Round(f) }

// floatToBigInt converts a whole-valued float64 to a *big.Int.
func floatToBigInt(f float64) *big.Int {
	bf := new(big.Float).SetFloat64(f)
	i, _ := bf.Int(nil)
	return i
}

// abs returns the absolute value of an int.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// String renders the Num exactly as Ruby's Kernel#inspect would for that kind:
// Integers bare ("5", "-2"), Rationals as "(n/d)", Floats via Ruby's float
// formatting (e.g. "2.0", "0.6", "-1.9999999999999998").
func (a Num) String() string {
	switch a.kind {
	case kindRat:
		return "(" + a.r.Num().String() + "/" + a.r.Denom().String() + ")"
	case kindFlt:
		return formatRubyFloat(a.f)
	default:
		return a.i.String()
	}
}

// formatRubyFloat formats f the way Ruby's Float#inspect does: the shortest
// decimal that round-trips, but always with a decimal point (so 2 prints as
// "2.0"), and the special Infinity / NaN spellings.
func formatRubyFloat(f float64) string {
	switch {
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case math.IsNaN(f):
		return "NaN"
	}
	s := strconvFormat(f)
	return s
}
