// Copyright (c) the go-ruby-matrix/matrix authors
//
// SPDX-License-Identifier: BSD-3-Clause

package matrix

import (
	"math"
	"strconv"
	"strings"
)

// strconvFormat renders a finite float64 the way MRI's Float#to_s / #inspect
// does. Go's strconv gives us the same shortest round-tripping digits, but its
// presentation differs (it drops a trailing ".0", and switches to scientific
// notation at different thresholds). Ruby's rule, from flo_to_s in numeric.c:
//
//   - let decpt be the position of the decimal point relative to the first
//     significant digit (the 'e' exponent of the shortest form, plus one);
//   - if -4 < decpt <= DBL_DIG (== 15) print in fixed notation, always with a
//     decimal point and at least one fractional digit ("2.0", "0.6");
//   - otherwise print in scientific notation with a single leading digit, a
//     mandatory fractional part, and a sign-and-two-digit exponent ("1.0e+20",
//     "1.0e-05").
func strconvFormat(f float64) string {
	if f == 0 {
		if math.Signbit(f) {
			return "-0.0"
		}
		return "0.0"
	}
	neg := math.Signbit(f)
	if neg {
		f = -f
	}

	// Shortest decimal digits and the base-10 exponent, via the 'e' form which
	// always yields "d.dddde±dd" with a single leading digit.
	es := strconv.FormatFloat(f, 'e', -1, 64)
	mantissa, expStr, _ := strings.Cut(es, "e")
	exp, _ := strconv.Atoi(expStr)
	digits := strings.Replace(mantissa, ".", "", 1) // significant digits, no point
	decpt := exp + 1                                // point sits after this many digits

	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}

	if decpt > -4 && decpt <= 15 {
		writeFixed(&b, digits, decpt)
	} else {
		writeSci(&b, digits, decpt)
	}
	return b.String()
}

// writeFixed emits the digits in fixed-point notation with the point placed
// after decpt significant digits, always leaving a fractional part.
func writeFixed(b *strings.Builder, digits string, decpt int) {
	switch {
	case decpt <= 0:
		// 0.000ddd  (|decpt| leading zeros after the point)
		b.WriteString("0.")
		b.WriteString(strings.Repeat("0", -decpt))
		b.WriteString(digits)
	case decpt >= len(digits):
		// integer-valued: ddd000.0
		b.WriteString(digits)
		b.WriteString(strings.Repeat("0", decpt-len(digits)))
		b.WriteString(".0")
	default:
		b.WriteString(digits[:decpt])
		b.WriteByte('.')
		b.WriteString(digits[decpt:])
	}
}

// writeSci emits "d.dddde±dd": one leading digit, a mandatory fractional part,
// and a signed two-digit (minimum) exponent.
func writeSci(b *strings.Builder, digits string, decpt int) {
	b.WriteByte(digits[0])
	b.WriteByte('.')
	if len(digits) > 1 {
		b.WriteString(digits[1:])
	} else {
		b.WriteByte('0')
	}
	b.WriteByte('e')
	e := decpt - 1
	if e < 0 {
		b.WriteByte('-')
		e = -e
	} else {
		b.WriteByte('+')
	}
	es := strconv.Itoa(e)
	if len(es) < 2 {
		b.WriteByte('0')
	}
	b.WriteString(es)
}
