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

// SCSI command definitions.

package scsismart

import "fmt"

// SCSI commands being used
const (
	SCSIInquiry        = 0x12
	SCSIModeSense6     = 0x1a
	SCSIReadCapacity10 = 0x25
	SCSIATAPassThru16  = 0x85

	// Minimum length of standard INQUIRY response
	INQRespLen = 36

	// SCSI-3 mode pages
	RigidDiskDriveGeometryPage = 0x04

	// Mode page control field
	ModePageControlDefault = 2
)

// SCSI CDB types
type CDB6 [6]byte
type CDB10 [10]byte
type CDB16 [16]byte

// InquiryResponse is the struct for SCSI INQUIRY response
type InquiryResponse struct {
	Peripheral byte
	_          byte
	Version    byte
	_          [5]byte
	VendorID   [8]byte
	ProductID  [16]byte
	ProductRev [4]byte
}

func (inquiry InquiryResponse) String() string {
	return fmt.Sprintf("%.8s  %.16s  %.4s", inquiry.VendorID, inquiry.ProductID, inquiry.ProductRev)
}
