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

package atasmart

import (
	"fmt"

	"github.com/openebs/smart/utilities"
)

// Table 47 of  T13/2161-D Revision 5
// See http://www.t13.org/Documents/UploadedDocuments/docs2013/d2161r5-ATAATAPI_Command_Set_-_3.pdf
var ataMinorVersions = map[uint16]string{
	0x0001: "ATA-1 X3T9.2/781D prior to revision 4",      // obsolete
	0x0002: "ATA-1 published, ANSI X3.221-1994",          // obsolete
	0x0003: "ATA-1 X3T9.2/781D revision 4",               // obsolete
	0x0004: "ATA-2 published, ANSI X3.279-1996",          // obsolete
	0x0005: "ATA-2 X3T10/948D prior to revision 2k",      // obsolete
	0x0006: "ATA-3 X3T10/2008D revision 1",               // obsolete
	0x0007: "ATA-2 X3T10/948D revision 2k",               // obsolete
	0x0008: "ATA-3 X3T10/2008D revision 0",               // obsolete
	0x0009: "ATA-2 X3T10/948D revision 3",                // obsolete
	0x000a: "ATA-3 published, ANSI X3.298-1997",          // obsolete
	0x000b: "ATA-3 X3T10/2008D revision 6",               // obsolete
	0x000c: "ATA-3 X3T13/2008D revision 7 and 7a",        // obsolete
	0x000d: "ATA/ATAPI-4 X3T13/1153D revision 6",         // obsolete
	0x000e: "ATA/ATAPI-4 T13/1153D revision 13",          // obsolete
	0x000f: "ATA/ATAPI-4 X3T13/1153D revision 7",         // obsolete
	0x0010: "ATA/ATAPI-4 T13/1153D revision 18",          // obsolete
	0x0011: "ATA/ATAPI-4 T13/1153D revision 15",          // obsolete
	0x0012: "ATA/ATAPI-4 published, ANSI NCITS 317-1998", // obsolete
	0x0013: "ATA/ATAPI-5 T13/1321D revision 3",
	0x0014: "ATA/ATAPI-4 T13/1153D revision 14", // obsolete
	0x0015: "ATA/ATAPI-5 T13/1321D revision 1",
	0x0016: "ATA/ATAPI-5 published, ANSI NCITS 340-2000",
	0x0017: "ATA/ATAPI-4 T13/1153D revision 17", // obsolete
	0x0018: "ATA/ATAPI-6 T13/1410D revision 0",
	0x0019: "ATA/ATAPI-6 T13/1410D revision 3a",
	0x001a: "ATA/ATAPI-7 T13/1532D revision 1",
	0x001b: "ATA/ATAPI-6 T13/1410D revision 2",
	0x001c: "ATA/ATAPI-6 T13/1410D revision 1",
	0x001d: "ATA/ATAPI-7 published, ANSI INCITS 397-2005",
	0x001e: "ATA/ATAPI-7 T13/1532D revision 0",
	0x001f: "ACS-3 T13/2161-D revision 3b",
	0x0021: "ATA/ATAPI-7 T13/1532D revision 4a",
	0x0022: "ATA/ATAPI-6 published, ANSI INCITS 361-2002",
	0x0027: "ATA8-ACS T13/1699-D revision 3c",
	0x0028: "ATA8-ACS T13/1699-D revision 6",
	0x0029: "ATA8-ACS T13/1699-D revision 4",
	0x0031: "ACS-2 T13/2015-D revision 2",
	0x0033: "ATA8-ACS T13/1699-D revision 3e",
	0x0039: "ATA8-ACS T13/1699-D revision 4c",
	0x0042: "ATA8-ACS T13/1699-D revision 3f",
	0x0052: "ATA8-ACS T13/1699-D revision 3b",
	0x005e: "ACS-4 T13/BSR INCITS 529 revision 5",
	0x006d: "ACS-3 T13/2161-D revision 5",
	0x0082: "ACS-2 published, ANSI INCITS 482-2012",
	0x0107: "ATA8-ACS T13/1699-D revision 2d",
	0x010a: "ACS-3 published, ANSI INCITS 522-2014",
	0x0110: "ACS-2 T13/2015-D revision 3",
	0x011b: "ACS-3 T13/2161-D revision 4",
}

// IdentDevData struct is an ATA IDENTIFY DEVICE struct. ATA8-ACS defines this as a page of 16-bit words.
type IdentDevData struct {
	_              [10]uint16  // ...
	SerialNumber   [20]byte    // Word 10..19, device serial number, padded with spaces (20h).
	_              [3]uint16   // ...
	FirmwareRev    [8]byte     // Word 23..26, device firmware revision, padded with spaces (20h).
	ModelNumber    [40]byte    // Word 27..46, device model number, padded with spaces (20h).
	_              [33]uint16  // ...
	MajorVer       uint16      // Word 80, major version number.
	MinorVer       uint16      // Word 81, minor version number.
	_              [3]uint16   // ...
	Word85         uint16      // Word 85, supported commands and feature sets.
	_              uint16      // ...
	Word87         uint16      // Word 87, supported commands and feature sets.
	_              [18]uint16  // ...
	SectorSize     uint16      // Word 106, Logical/physical sector size.
	_              [1]uint16   // ...
	WWN            [4]uint16   // Word 108..111, WWN (World Wide Name).
	_              [105]uint16 // ...
	RotationRate   uint16      // Word 217, nominal media rotation rate.
	_              [4]uint16   // ...
	TransportMajor uint16      // Word 222, transport major version number.
	_              [33]uint16  // ...
} // 512 bytes

// swapByteOrder swaps the order of every second byte in a byte slice (modifies slice in-place).
func (d *IdentDevData) swapByteOrder(b []byte) []byte {
	tmp := make([]byte, len(b))

	for i := 0; i < len(b); i += 2 {
		tmp[i], tmp[i+1] = b[i+1], b[i]
	}

	return tmp
}

// GetSerialNumber returns the serial number of a device from an ATA IDENTIFY command.
func (d *IdentDevData) GetSerialNumber() []byte {
	return d.swapByteOrder(d.SerialNumber[:])
}

// GetModelNumber returns the model number of a device from an ATA IDENTIFY command.
func (d *IdentDevData) GetModelNumber() []byte {
	return d.swapByteOrder(d.ModelNumber[:])
}

// GetFirmwareRevision returns the firmware version of a device from an ATA IDENTIFY command.
func (d *IdentDevData) GetFirmwareRevision() []byte {
	return d.swapByteOrder(d.FirmwareRev[:])
}

// GetWWN returns the worldwide unique name for a disk
func (d *IdentDevData) GetWWN() string {
	NAA := d.WWN[0] >> 12
	IEEEOUI := (uint32(d.WWN[0]&0x0fff) << 12) | (uint32(d.WWN[1]) >> 4)
	UniqueID := ((uint64(d.WWN[1]) & 0xf) << 32) | (uint64(d.WWN[2]) << 16) | uint64(d.WWN[3])

	return fmt.Sprintf("%x %06x %09x", NAA, IEEEOUI, UniqueID)
}

// GetSectorSize returns logical and physical sector sizes of a disk
func (d *IdentDevData) GetSectorSize() (uint16, uint16) {
	var (
		LogSec, PhySec uint16 = 512, 512
	)
	if (d.SectorSize & 0xc000) == 0x4000 {
		if (d.SectorSize & 0x2000) != 0x0000 {
			// Physical sector size is multiple of logical sector size
			PhySec <<= (d.SectorSize & 0x0f)
		}
	}
	return LogSec, PhySec
}

// GetATAMajorVersion returns the ATA major version from an ATA IDENTIFY command.
func (d *IdentDevData) GetATAMajorVersion() (s string) {
	if (d.MajorVer == 0) || (d.MajorVer == 0xffff) {
		s = "This device does not report ATA major version"
		return
	}

	switch utilities.MSignificantBit(uint(d.MajorVer)) {
	case 1:
		s = "ATA-1"
	case 2:
		s = "ATA-2"
	case 3:
		s = "ATA-3"
	case 4:
		s = "ATA/ATAPI-4"
	case 5:
		s = "ATA/ATAPI-5"
	case 6:
		s = "ATA/ATAPI-6"
	case 7:
		s = "ATA/ATAPI-7"
	case 8:
		s = "ATA8-ACS"
	case 9:
		s = "ACS-2"
	case 10:
		s = "ACS-3"
	}

	return
}

// GetATAMinorVersion returns the ATA minor version from an ATA IDENTIFY command.
func (d *IdentDevData) GetATAMinorVersion() string {
	if (d.MinorVer == 0) || (d.MinorVer == 0xffff) {
		return "This device does not report ATA minor version"
	}

	// Since the ATA minor version word is not a bitmask, we simply do a map lookup
	if s, ok := ataMinorVersions[d.MinorVer]; ok {
		return s
	}

	return "unknown"
}

// Transport returns the type of ata transport being used such as serial ATA, parallel ATA.
func (d *IdentDevData) Transport() (s string) {
	if (d.TransportMajor == 0) || (d.TransportMajor == 0xffff) {
		s = "This device does not report transport"
		return
	}

	switch d.TransportMajor >> 12 {
	case 0x0:
		s = "Parallel ATA"
	case 0x1:
		s = "Serial ATA"

		switch utilities.MSignificantBit(uint(d.TransportMajor & 0x0fff)) {
		case 0:
			s += " ATA8-AST"
		case 1:
			s += " SATA 1.0a"
		case 2:
			s += " SATA II Ext"
		case 3:
			s += " SATA 2.5"
		case 4:
			s += " SATA 2.6"
		case 5:
			s += " SATA 3.0"
		case 6:
			s += " SATA 3.1"
		case 7:
			s += " SATA 3.2"
		default:
			s += fmt.Sprintf(" SATA (%#03x)", d.TransportMajor&0x0fff)
		}
	case 0xe:
		s = fmt.Sprintf("PCIe (%#03x)", d.TransportMajor&0x0fff)
	default:
		s = fmt.Sprintf("Unknown (%#04x)", d.TransportMajor)
	}

	return
}
