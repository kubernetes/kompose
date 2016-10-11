// Copyright ©2013 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mat64

import (
	"fmt"
	"strconv"
)

// Format prints a pretty representation of m to the fs io.Writer. The format character c
// specifies the numerical representation of of elements; valid values are those for float64
// specified in the fmt package, with their associated flags. In addition to this, a '#' for
// all valid verbs except 'v' indicates that zero values be represented by the dot character.
// The '#' associated with the 'v' verb formats the matrix with Go syntax representation.
// The printed range of the matrix can be limited by specifying a positive value for margin;
// If margin is greater than zero, only the first and last margin rows/columns of the matrix
// are output.
func Format(m Matrix, margin int, dot byte, fs fmt.State, c rune) {
	rows, cols := m.Dims()

	var printed int
	if margin <= 0 {
		printed = rows
		if cols > printed {
			printed = cols
		}
	} else {
		printed = margin
	}

	prec, pOk := fs.Precision()
	if !pOk {
		prec = -1
	}

	var (
		maxWidth int
		buf, pad []byte
	)
	switch c {
	case 'v', 'e', 'E', 'f', 'F', 'g', 'G':
		// Note that the '#' flag should have been dealt with by the type.
		// So %v is treated exactly as %g here.
		if c == 'v' {
			buf, maxWidth = maxCellWidth(m, 'g', printed, prec)
		} else {
			buf, maxWidth = maxCellWidth(m, c, printed, prec)
		}
	default:
		fmt.Fprintf(fs, "%%!%c(%T=Dims(%d, %d))", c, m, rows, cols)
		return
	}
	width, _ := fs.Width()
	width = max(width, maxWidth)
	pad = make([]byte, max(width, 2))
	for i := range pad {
		pad[i] = ' '
	}

	if rows > 2*printed || cols > 2*printed {
		fmt.Fprintf(fs, "Dims(%d, %d)\n", rows, cols)
	}

	skipZero := fs.Flag('#')
	for i := 0; i < rows; i++ {
		var el string
		switch {
		case rows == 1:
			fmt.Fprint(fs, "[")
			el = "]"
		case i == 0:
			fmt.Fprint(fs, "⎡")
			el = "⎤\n"
		case i < rows-1:
			fmt.Fprint(fs, "⎢")
			el = "⎥\n"
		default:
			fmt.Fprint(fs, "⎣")
			el = "⎦"
		}

		for j := 0; j < cols; j++ {
			if j >= printed && j < cols-printed {
				j = cols - printed - 1
				if i == 0 || i == rows-1 {
					fmt.Fprint(fs, "...  ...  ")
				} else {
					fmt.Fprint(fs, "          ")
				}
				continue
			}

			v := m.At(i, j)
			if v == 0 && skipZero {
				buf = buf[:1]
				buf[0] = dot
			} else {
				if c == 'v' {
					buf = strconv.AppendFloat(buf[:0], v, 'g', prec, 64)
				} else {
					buf = strconv.AppendFloat(buf[:0], v, byte(c), prec, 64)
				}
			}
			if fs.Flag('-') {
				fs.Write(buf)
				fs.Write(pad[:width-len(buf)])
			} else {
				fs.Write(pad[:width-len(buf)])
				fs.Write(buf)
			}

			if j < cols-1 {
				fs.Write(pad[:2])
			}
		}

		fmt.Fprint(fs, el)

		if i >= printed-1 && i < rows-printed && 2*printed < rows {
			i = rows - printed - 1
			fmt.Fprint(fs, " .\n .\n .\n")
			continue
		}
	}
}

func maxCellWidth(m Matrix, c rune, printed, prec int) ([]byte, int) {
	var (
		buf        = make([]byte, 0, 64)
		rows, cols = m.Dims()
		max        int
	)
	for i := 0; i < rows; i++ {
		if i >= printed-1 && i < rows-printed && 2*printed < rows {
			i = rows - printed - 1
			continue
		}
		for j := 0; j < cols; j++ {
			if j >= printed && j < cols-printed {
				continue
			}

			buf = strconv.AppendFloat(buf, m.At(i, j), byte(c), prec, 64)
			if len(buf) > max {
				max = len(buf)
			}
			buf = buf[:0]
		}
	}
	return buf, max
}
