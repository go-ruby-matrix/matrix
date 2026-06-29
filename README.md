<p align="center"><img src="https://raw.githubusercontent.com/go-ruby-matrix/brand/main/social/go-ruby-matrix-matrix.png" alt="go-ruby-matrix/matrix" width="720"></p>

# matrix — go-ruby-matrix

[![Docs](https://img.shields.io/badge/docs-mkdocs--material-DC2626)](https://go-ruby-matrix.github.io/docs/)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26.4%2B-00ADD8)](https://go.dev/dl/)
[![Coverage](https://img.shields.io/badge/coverage-100%25-1a7f37)](#tests--coverage)

**A pure-Go (no cgo) reimplementation of Ruby's [`matrix`](https://docs.ruby-lang.org/en/master/Matrix.html)
standard library** — `Matrix` and `Vector` with MRI 4.0.5-faithful behaviour. It
matches MRI's `to_s` / `inspect` formatting, its error classes, and — crucially —
its **exact arithmetic**: an Integer matrix's `determinant` stays an Integer, and
`inverse` produces exact Rationals (`Matrix[[1,2],[3,4]].inverse` →
`Matrix[[(-2/1), (1/1)], [(3/2), (-1/2)]]`) — **without any Ruby runtime**.

It is the `matrix` backend for
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby), but is a
**standalone, reusable** module with no dependency on the Ruby runtime — a sibling
of [go-ruby-regexp](https://github.com/go-ruby-regexp/regexp) (the Onigmo engine),
[go-ruby-erb](https://github.com/go-ruby-erb/erb) (the ERB compiler) and
[go-ruby-yaml](https://github.com/go-ruby-yaml/yaml) (the Psych port).

> **Exact where MRI is exact.** Ruby's `matrix` carries Ruby's numeric tower:
> Integer and Rational entries stay exact, and a division only escapes to Float
> when the input is Float. This package mirrors that with a `Num` tower built on
> `math/big` — `*big.Int` for Integers, `*big.Rat` for Rationals, `float64` for
> Floats — and reproduces MRI's `inverse_from` Gauss-Jordan **step for step**
> (partial pivoting on the largest absolute pivot), so even a Float matrix's
> inverse reproduces MRI's exact rounding (e.g. `-1.9999999999999998`).

## Features

Faithful port of the `Matrix` and `Vector` API, validated against the `ruby`
binary on every supported platform:

- **Constructors** — `New` / `Build` / `Identity` / `Zero` / `Diagonal` /
  `Scalar` / `RowVector` / `ColumnVector` / `Rows` / `Columns` / `HStack` /
  `VStack`.
- **Accessors** — `RowCount` / `ColumnCount` / `At` / `Row` / `Column` / `Each` /
  `EachWithIndex` / `Minor` / `FirstMinor` / `ToA`.
- **Operations** — `Add` / `Sub` / `Neg` / `Mul` (matrix·matrix, ·scalar,
  ·vector) / `Div` / `Pow` / `Transpose` / `Trace` / `Determinant` / `Inverse` /
  `Rank` / `RoundEntries` — all exact over the numeric tower.
- **Predicates** — `Square` / `IsDiagonal` / `Symmetric` / `Orthogonal` /
  `Singular` / `Regular` / `LowerTriangular` / `UpperTriangular` / `IsZero`.
- **Equality & formatting** — `Eql` (`==`), `Hash`, `ToS` / `Inspect`
  (`Matrix[[…], […]]`, `Matrix.empty(r, c)`).
- **Vector** — `Elements` / `At` / `Size` / `Add` / `Sub` / `Mul` /
  `InnerProduct` (dot) / `CrossProduct` / `Magnitude` (norm) / `Normalize` /
  `Each` / `Map` / `Angle` / `Eql` / `Independent`.
- **Errors** — `ErrDimensionMismatch`, `ErrNotRegular`, `ErrOperationNotDefined`,
  with MRI's messages.

CGO-free, dependency-free, **100% test coverage**, `gofmt` + `go vet` clean, and
green across the six 64-bit Go targets (amd64, arm64, riscv64, loong64, ppc64le,
s390x) and three OSes (Linux, macOS, Windows).

## Install

```sh
go get github.com/go-ruby-matrix/matrix
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/go-ruby-matrix/matrix"
)

func main() {
	m, _ := matrix.New([][]any{{1, 2}, {3, 4}})

	d, _ := m.Determinant()
	fmt.Println(d)             // -2          (Integer — exact)

	inv, _ := m.Inverse()
	fmt.Println(inv.ToS())     // Matrix[[(-2/1), (1/1)], [(3/2), (-1/2)]]  (exact Rationals)

	p2, _ := m.Pow(2)
	fmt.Println(p2.ToS())      // Matrix[[7, 10], [15, 22]]

	v1 := mustVec(1, 2, 3)
	v2 := mustVec(4, 5, 6)
	dot, _ := v1.InnerProduct(v2)
	fmt.Println(dot)           // 32

	cross, _ := mustVec(1, 0, 0).CrossProduct(mustVec(0, 1, 0))
	fmt.Println(cross.ToS())   // Vector[0, 0, 1]

	fmt.Println(mustVec(3, 4).Magnitude()) // 5.0
}

func mustVec(xs ...any) *matrix.Vector { v, _ := matrix.NewVector(xs); return v }
```

## The numeric tower (exact arithmetic)

Entries flow through `matrix.Num`, the slice of Ruby's numeric tower the `matrix`
library depends on. A host (rbgo) maps its own numerics to and from it:

| Ruby       | Go (constructors accept)        | `Num` kind | `String()` form          |
| ---------- | ------------------------------- | ---------- | ------------------------ |
| `Integer`  | `int…`, `uint…`, `*big.Int`     | Integer    | `5`, `-2`                |
| `Rational` | `*big.Rat`                      | Rational   | `(3/2)`, `(0/1)`         |
| `Float`    | `float32`, `float64`            | Float      | `2.0`, `0.6`, `Infinity` |

Promotion follows Ruby: Integer op Integer stays Integer; any Rational (no Float)
yields a Rational that **stays** Rational even when whole (so `inverse` prints
`(1/1)`, not `1`); any Float dominates. Exact division uses `Num.Quo` (Ruby's
`Integer#quo`), which turns Integer/Integer into a Rational — exactly how MRI's
`inverse`, `normalize` and matrix `/` divide.

`Determinant` uses cofactor (Laplace) expansion with only `+ − ×`, so an Integer
matrix yields an Integer determinant; `Inverse` and `Rank` run exact Gaussian
elimination over the tower.

## API

```go
// Matrix
func New(rows [][]any) (*Matrix, error)
func Build(r, c int, fn func(i, j int) any) (*Matrix, error)
func Identity(n int) *Matrix
func Zero(r, c int) *Matrix
func Diagonal(values ...any) (*Matrix, error)
func Scalar(n int, value any) (*Matrix, error)
func RowVector(values []any) (*Matrix, error)
func ColumnVector(values []any) (*Matrix, error)
func Rows(rows [][]any) (*Matrix, error)
func Columns(cols [][]any) (*Matrix, error)
func HStack(ms ...*Matrix) (*Matrix, error)
func VStack(ms ...*Matrix) (*Matrix, error)

func (m *Matrix) RowCount() int
func (m *Matrix) ColumnCount() int
func (m *Matrix) At(i, j int) (Num, bool)
func (m *Matrix) Row(i int) (*Vector, bool)
func (m *Matrix) Column(j int) (*Vector, bool)
func (m *Matrix) Each(fn func(v Num))
func (m *Matrix) EachWithIndex(fn func(v Num, i, j int))
func (m *Matrix) ToA() [][]Num
func (m *Matrix) Minor(r0, r1, c0, c1 int) (*Matrix, error)
func (m *Matrix) FirstMinor(i, j int) (*Matrix, error)

func (m *Matrix) Add(other *Matrix) (*Matrix, error)
func (m *Matrix) Sub(other *Matrix) (*Matrix, error)
func (m *Matrix) Neg() *Matrix
func (m *Matrix) Mul(other *Matrix) (*Matrix, error)
func (m *Matrix) MulScalar(scalar any) (*Matrix, error)
func (m *Matrix) MulVector(v *Vector) (*Vector, error)
func (m *Matrix) Div(other *Matrix) (*Matrix, error)
func (m *Matrix) DivScalar(scalar any) (*Matrix, error)
func (m *Matrix) Pow(p int) (*Matrix, error)
func (m *Matrix) Transpose() *Matrix
func (m *Matrix) Trace() (Num, error)
func (m *Matrix) Determinant() (Num, error)
func (m *Matrix) Inverse() (*Matrix, error)
func (m *Matrix) Rank() int
func (m *Matrix) RoundEntries(ndigits int) *Matrix

func (m *Matrix) Square() bool
func (m *Matrix) IsZero() bool
func (m *Matrix) IsDiagonal() bool
func (m *Matrix) Symmetric() bool
func (m *Matrix) LowerTriangular() bool
func (m *Matrix) UpperTriangular() bool
func (m *Matrix) Singular() (bool, error)
func (m *Matrix) Regular() (bool, error)
func (m *Matrix) Orthogonal() (bool, error)
func (m *Matrix) Eql(other *Matrix) bool
func (m *Matrix) Hash() uint64
func (m *Matrix) ToS() string
func (m *Matrix) Inspect() string

// Vector
func NewVector(values []any) (*Vector, error)
func Independent(vs ...*Vector) (bool, error)
func (v *Vector) Elements() []Num
func (v *Vector) Size() int
func (v *Vector) At(i int) (Num, bool)
func (v *Vector) Each(fn func(n Num))
func (v *Vector) Map(fn func(n Num) any) (*Vector, error)
func (v *Vector) Add(other *Vector) (*Vector, error)
func (v *Vector) Sub(other *Vector) (*Vector, error)
func (v *Vector) Mul(scalar any) (*Vector, error)
func (v *Vector) InnerProduct(other *Vector) (Num, error)
func (v *Vector) CrossProduct(other *Vector) (*Vector, error)
func (v *Vector) Magnitude() Num
func (v *Vector) Normalize() (*Vector, error)
func (v *Vector) Angle(other *Vector) (Num, error)
func (v *Vector) Eql(other *Vector) bool
func (v *Vector) ToS() string
func (v *Vector) Inspect() string

// Num (numeric tower)
func NewInt(v int64) Num
func NewBigInt(v *big.Int) Num
func NewRat(n, d int64) Num
func NewBigRat(v *big.Rat) Num
func NewFloat(v float64) Num

// Errors (mirror ExceptionForMatrix)
var ErrDimensionMismatch   // "Dimension mismatch"
var ErrNotRegular          // "Not Regular Matrix"
var ErrOperationNotDefined
```

## Tests & coverage

The suite pairs deterministic, ruby-free tests (which alone hold coverage at
100%, so the qemu cross-arch and Windows lanes pass the gate) with a **differential
MRI oracle**: every operation's `ToS` is compared against the system `ruby`'s
`-rmatrix … .inspect`, including the exact-arithmetic results (Integer
determinant, Rational inverse, Float-matrix inverse rounding) and the
Ruby-faithful float formatter. The oracle binmodes stdin/stdout so Windows
text-mode never pollutes the bytes, gates itself on `RUBY_VERSION >= "4.0"`, and
skips where `ruby` is absent.

```sh
COVERPKG=$(go list ./... | paste -sd, -)
go test -race -coverpkg="$COVERPKG" -coverprofile=cover.out ./...
go tool cover -func=cover.out | tail -1   # 100.0%
```

## License

BSD-3-Clause — see [LICENSE](LICENSE). Copyright the go-ruby-matrix/matrix authors.
