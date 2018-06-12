/*
Copyright 2018 The OpenEBS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Miscellaneous utility functions

package utilities

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"unsafe"
)

var (
	NativeEndian binary.ByteOrder
)

// Determine native endianness of system
func init() {
	i := uint32(1)
	b := (*[4]byte)(unsafe.Pointer(&i))
	if b[0] == 1 {
		NativeEndian = binary.LittleEndian
	} else {
		NativeEndian = binary.BigEndian
	}
}

// MSignificantBit finds the most significant bit set in a uint
func MSignificantBit(x uint) int {
	if x == 0 {
		return 0
	}

	return bits.Len(x) - 1
}

// ConvertBytes converts a uint64 byte quantity using human-readble units, e.g. kilobyte, megabyte.
func ConvertBytes(v uint64) string {
	var i int

	suffixes := [...]string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	d := uint64(1)

	for i = 0; i < len(suffixes)-1; i++ {
		if v >= d*1000 {
			d *= 1000
		} else {
			break
		}
	}

	if i == 0 {
		return fmt.Sprintf("%d %s", v, suffixes[i])
	} else {
		// Print upto 3 significant digits
		return fmt.Sprintf("%.3g %s", float64(v)/float64(d), suffixes[i])
	}
}
