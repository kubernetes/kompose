// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mat64

import (
	"encoding/binary"
)

var (
	littleEndian  = binary.LittleEndian
	bigEndian     = binary.BigEndian
	defaultEndian = littleEndian

	sizeInt64   = binary.Size(int64(0))
	sizeFloat64 = binary.Size(float64(0))
)
