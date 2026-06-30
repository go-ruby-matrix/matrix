// Copyright (c) the go-ruby-matrix/matrix authors
//
// SPDX-License-Identifier: BSD-3-Clause

package matrix

import (
	"os/exec"
	"strings"
	"testing"
)

// rubyBin locates a usable `ruby` whose stdlib `matrix` is the 4.0-era gem
// (RUBY_VERSION >= "4.0"). The oracle tests skip themselves when ruby is absent
// (the qemu cross-arch lanes, the Windows lane) or too old, so the
// deterministic suite alone drives the 100% gate everywhere.
func rubyBin(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("ruby")
	if err != nil {
		t.Skip("ruby not on PATH; skipping MRI oracle")
	}
	out, err := exec.Command(path, "-e", "print(RUBY_VERSION >= \"4.0\")").CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		t.Skipf("ruby %s < 4.0; skipping MRI oracle", strings.TrimSpace(string(out)))
	}
	return path
}

// rubyMatrix evaluates expr under `ruby -rmatrix` and returns its inspected
// value with the trailing newline trimmed. The preamble binmodes stdin/stdout so
// Windows text-mode never rewrites the bytes (the go-ruby-erb lesson).
func rubyMatrix(t *testing.T, bin, expr string) string {
	t.Helper()
	script := "$stdout.binmode\n$stdin.binmode\nprint((" + expr + ").inspect)\n"
	out, err := exec.Command(bin, "-rmatrix", "-e", script).CombinedOutput()
	if err != nil {
		t.Fatalf("ruby error: %v\nexpr: %s\noutput:\n%s", err, expr, out)
	}
	return strings.TrimRight(string(out), "\n")
}

// TestOracleMatrix checks this package's ToS against MRI's `inspect` for the
// full range of Matrix operations — including the exact-arithmetic results
// (Integer determinant, Rational inverse, Float-matrix inverse rounding).
func TestOracleMatrix(t *testing.T) {
	bin := rubyBin(t)
	cases := []struct {
		name string
		ruby string
		got  func(t *testing.T) string
	}{
		{"new", "Matrix[[1,2],[3,4]]", func(t *testing.T) string { return mat(t, [][]any{{1, 2}, {3, 4}}).ToS() }},
		{"identity", "Matrix.identity(3)", func(t *testing.T) string { return Identity(3).ToS() }},
		{"zero", "Matrix.zero(2,3)", func(t *testing.T) string { return Zero(2, 3).ToS() }},
		{"diagonal", "Matrix.diagonal(1,2,3)", func(t *testing.T) string { m, _ := Diagonal(1, 2, 3); return m.ToS() }},
		{"scalar", "Matrix.scalar(3,5)", func(t *testing.T) string { m, _ := Scalar(3, 5); return m.ToS() }},
		{"build", "Matrix.build(2,2){|i,j| i*2+j}", func(t *testing.T) string {
			m, _ := Build(2, 2, func(i, j int) any { return i*2 + j })
			return m.ToS()
		}},
		{"transpose", "Matrix[[1,2],[3,4]].transpose", func(t *testing.T) string { return mat(t, [][]any{{1, 2}, {3, 4}}).Transpose().ToS() }},
		{"add", "Matrix[[1,2],[3,4]] + Matrix[[5,6],[7,8]]", func(t *testing.T) string {
			r, _ := mat(t, [][]any{{1, 2}, {3, 4}}).Add(mat(t, [][]any{{5, 6}, {7, 8}}))
			return r.ToS()
		}},
		{"mul", "Matrix[[1,2],[3,4]] * Matrix[[2,0],[1,2]]", func(t *testing.T) string {
			r, _ := mat(t, [][]any{{1, 2}, {3, 4}}).Mul(mat(t, [][]any{{2, 0}, {1, 2}}))
			return r.ToS()
		}},
		{"scalarmul", "Matrix[[1,2],[3,4]] * 3", func(t *testing.T) string {
			r, _ := mat(t, [][]any{{1, 2}, {3, 4}}).MulScalar(3)
			return r.ToS()
		}},
		{"pow", "Matrix[[1,2],[3,4]] ** 3", func(t *testing.T) string { r, _ := mat(t, [][]any{{1, 2}, {3, 4}}).Pow(3); return r.ToS() }},
		{"det2", "Matrix[[1,2],[3,4]].determinant", func(t *testing.T) string {
			d, _ := mat(t, [][]any{{1, 2}, {3, 4}}).Determinant()
			return d.String()
		}},
		{"det3", "Matrix[[1,2,3],[4,5,6],[7,8,10]].determinant", func(t *testing.T) string {
			d, _ := mat(t, [][]any{{1, 2, 3}, {4, 5, 6}, {7, 8, 10}}).Determinant()
			return d.String()
		}},
		{"inverse2", "Matrix[[1,2],[3,4]].inverse", func(t *testing.T) string {
			r, _ := mat(t, [][]any{{1, 2}, {3, 4}}).Inverse()
			return r.ToS()
		}},
		{"inverse3", "Matrix[[1,2,3],[0,1,4],[5,6,0]].inverse", func(t *testing.T) string {
			r, _ := mat(t, [][]any{{1, 2, 3}, {0, 1, 4}, {5, 6, 0}}).Inverse()
			return r.ToS()
		}},
		{"floatinverse", "Matrix[[1.0,2.0],[3.0,4.0]].inverse", func(t *testing.T) string {
			r, _ := mat(t, [][]any{{1.0, 2.0}, {3.0, 4.0}}).Inverse()
			return r.ToS()
		}},
		{"floatdet", "Matrix[[1.0,2.0],[3.0,4.0]].determinant", func(t *testing.T) string {
			d, _ := mat(t, [][]any{{1.0, 2.0}, {3.0, 4.0}}).Determinant()
			return d.String()
		}},
		{"rank", "Matrix[[1,2,3],[2,4,6],[1,0,1]].rank", func(t *testing.T) string {
			return itoa(mat(t, [][]any{{1, 2, 3}, {2, 4, 6}, {1, 0, 1}}).Rank())
		}},
		{"trace", "Matrix[[1,2],[3,4]].trace", func(t *testing.T) string {
			tr, _ := mat(t, [][]any{{1, 2}, {3, 4}}).Trace()
			return tr.String()
		}},
		{"round", "Matrix[[1.234,2.567]].round(1)", func(t *testing.T) string {
			return mat(t, [][]any{{1.234, 2.567}}).RoundEntries(1).ToS()
		}},
		// round with no argument (== round(0)) yields Integer entries in MRI.
		{"round0", "Matrix[[1.4,2.6]].round", func(t *testing.T) string {
			return mat(t, [][]any{{1.4, 2.6}}).RoundEntries(0).ToS()
		}},
		// round with a negative argument rounds at the 10**(-n) place, Integer.
		{"roundneg", "Matrix[[14.5,25.5]].round(-1)", func(t *testing.T) string {
			return mat(t, [][]any{{14.5, 25.5}}).RoundEntries(-1).ToS()
		}},
		// no-arg round of Rational entries → Integer, half away from zero.
		{"roundrat", "Matrix[[Rational(3,2),Rational(7,2)]].round", func(t *testing.T) string {
			return mat(t, [][]any{{NewRat(3, 2), NewRat(7, 2)}}).RoundEntries(0).ToS()
		}},
		// round(n>=1) keeps the Rational kind.
		{"roundratkeep", "Matrix[[Rational(7,3)]].round(1)", func(t *testing.T) string {
			return mat(t, [][]any{{NewRat(7, 3)}}).RoundEntries(1).ToS()
		}},
		{"div", "Matrix[[1,2],[3,4]] / Matrix.identity(2)", func(t *testing.T) string {
			r, _ := mat(t, [][]any{{1, 2}, {3, 4}}).Div(Identity(2))
			return r.ToS()
		}},
		// Matrix#/ with an Integer scalar floors per Integer/Integer entry.
		{"divscalarint", "Matrix[[3,5],[7,9]] / 2", func(t *testing.T) string {
			r, _ := mat(t, [][]any{{3, 5}, {7, 9}}).DivScalar(2)
			return r.ToS()
		}},
		// Negative numerator floors toward negative infinity.
		{"divscalarintneg", "Matrix[[-3,5]] / 2", func(t *testing.T) string {
			r, _ := mat(t, [][]any{{-3, 5}}).DivScalar(2)
			return r.ToS()
		}},
		// A Float scalar makes every entry a Float.
		{"divscalarfloat", "Matrix[[3,5],[7,9]] / 2.0", func(t *testing.T) string {
			r, _ := mat(t, [][]any{{3, 5}, {7, 9}}).DivScalar(2.0)
			return r.ToS()
		}},
		// A Rational scalar keeps entries Rational.
		{"divscalarrat", "Matrix[[3,5]] / Rational(2,1)", func(t *testing.T) string {
			r, _ := mat(t, [][]any{{3, 5}}).DivScalar(NewRat(2, 1))
			return r.ToS()
		}},
		{"empty", "Matrix.empty(2,0)", func(t *testing.T) string { return Zero(2, 0).ToS() }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			want := rubyMatrix(t, bin, c.ruby)
			if got := c.got(t); got != want {
				t.Errorf("%s: got %q, MRI %q", c.name, got, want)
			}
		})
	}
}

// TestOracleVector mirrors TestOracleMatrix for the Vector API.
func TestOracleVector(t *testing.T) {
	bin := rubyBin(t)
	cases := []struct {
		name string
		ruby string
		got  func(t *testing.T) string
	}{
		{"new", "Vector[1,2,3]", func(t *testing.T) string { return vec(t, 1, 2, 3).ToS() }},
		{"add", "Vector[1,2] + Vector[3,4]", func(t *testing.T) string { r, _ := vec(t, 1, 2).Add(vec(t, 3, 4)); return r.ToS() }},
		{"sub", "Vector[1,2] - Vector[3,4]", func(t *testing.T) string { r, _ := vec(t, 1, 2).Sub(vec(t, 3, 4)); return r.ToS() }},
		{"scalarmul", "Vector[1,2] * 3", func(t *testing.T) string { r, _ := vec(t, 1, 2).Mul(3); return r.ToS() }},
		{"dot", "Vector[1,2,3].inner_product(Vector[4,5,6])", func(t *testing.T) string {
			d, _ := vec(t, 1, 2, 3).InnerProduct(vec(t, 4, 5, 6))
			return d.String()
		}},
		{"cross", "Vector[1,0,0].cross_product(Vector[0,1,0])", func(t *testing.T) string {
			r, _ := vec(t, 1, 0, 0).CrossProduct(vec(t, 0, 1, 0))
			return r.ToS()
		}},
		{"magnitude", "Vector[3,4].magnitude", func(t *testing.T) string { return vec(t, 3, 4).Magnitude().String() }},
		{"normalize", "Vector[3,4].normalize", func(t *testing.T) string { r, _ := vec(t, 3, 4).Normalize(); return r.ToS() }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			want := rubyMatrix(t, bin, c.ruby)
			if got := c.got(t); got != want {
				t.Errorf("%s: got %q, MRI %q", c.name, got, want)
			}
		})
	}
}

// TestOracleFloatFormat cross-checks the Ruby-faithful float formatter against
// MRI's Float#inspect for an awkward spread of magnitudes (the scientific /
// fixed boundary, sub-normal-ish small values, whole floats).
func TestOracleFloatFormat(t *testing.T) {
	bin := rubyBin(t)
	values := []float64{
		2.0, 0.6, 5.0, 0.0001, 1e-5, 1e15, 1e16, 1e20, 1e21,
		123456789012345.6, 9.999999999999998e+15, 1.5e-10, 1234567.89,
	}
	for _, f := range values {
		// Hand MRI the exact bits via the shortest decimal we produce, so both
		// sides format the *same* float64.
		lit := strconvFormat(f)
		want := rubyMatrix(t, bin, "Float("+strings.NewReplacer("Infinity", "1.0/0", "NaN", "0.0/0").Replace(lit)+")")
		if got := formatRubyFloat(f); got != want {
			t.Errorf("formatRubyFloat(%v) = %q, MRI %q", f, got, want)
		}
	}
}
