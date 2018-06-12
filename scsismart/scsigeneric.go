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

// SCSI generic IO functions.

package scsismart

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/openebs/smart/ioctl"
	"github.com/openebs/smart/utilities"
)

// SCSI generic (sg)
// See dxfer_direction http://sg.danny.cz/sg/p/sg_v3_ho.html
const (
	SGDxferNone      = -1 //SCSI Test Unit Ready command
	SGDxferToDev     = -2 //SCSI WRITE command
	SGDxferFromDev   = -3 //SCSI READ command
	SGDxferToFromDev = -4

	SGInfoOkMask = 0x1
	SGInfoOk     = 0x0

	SGIO = 0x2285

	// DefaultTimeout in millisecs
	DefaultTimeout = 20000
)

// sg_io_hdr_t structure See http://sg.danny.cz/sg/p/sg_v3_ho.html
type sgIOHeader struct {
	interfaceID    int32   // 'S' for SCSI generic (required)
	dxferDirection int32   // data transfer direction
	cmdLen         uint8   // SCSI command length (<= 16 bytes)
	mxSBLen        uint8   // max length to write to sbp
	iovecCount     uint16  // 0 implies no scatter gather
	dxferLen       uint32  // byte count of data transfer
	dxferp         uintptr // points to data transfer memory or scatter gather list
	cmdp           uintptr // points to command to perform
	sbp            uintptr // points to sense_buffer memory
	timeout        uint32  // MAX_UINT -> no timeout (unit: millisec)
	flags          uint32  // 0 -> default, see SG_FLAG...
	packID         int32   // unused internally (normally)
	usrPtr         uintptr // unused internally
	status         uint8   // SCSI status
	maskedStatus   uint8   // shifted, masked scsi status
	msgStatus      uint8   // messaging level data (optional)
	SBLenwr        uint8   // byte count actually written to sbp
	hostStatus     uint16  // errors from host adapter
	driverStatus   uint16  // errors from software driver
	resid          int32   // dxfer_len - actual_transferred
	duration       uint32  // time taken by cmd (unit: millisec)
	info           uint32  // auxiliary information
}

type sgIOErr struct {
	scsiStatus   uint8
	hostStatus   uint16
	driverStatus uint16
	senseBuf     [32]byte
}

// DiskAttr is the structure for returning disk details
type DiskAttr struct {
	SCSIInquiry      InquiryResponse
	VendorID         uint16
	UserCapacity     uint64
	LBSize           uint16
	PBSize           uint16
	SerialNumber     string
	LuWWNDeviceID    string
	FirmwareRevision string
	ModelNumber      string
	RotationRate     uint16
	ATAMajorVersion  string
	ATAMinorVersion  string
	Transport        string
}

func (e sgIOErr) Error() string {
	return fmt.Sprintf("SCSI status: %#02x, host status: %#02x, driver status: %#02x",
		e.scsiStatus, e.hostStatus, e.driverStatus)
}

// Dev is the top-level device interface. All supported device types must implement these methods.
type Dev interface {
	Open() error
	Close() error
	PrintDiskInfo() error
	GetDiskInfo() (DiskAttr, error)
}

// SCSIDevice structure
type SCSIDevice struct {
	Name string
	fd   int
}

// DetectSCSIType returns the type of SCSI device
func DetectSCSIType(name string) (Dev, error) {
	dev := SCSIDevice{Name: name}

	if err := dev.Open(); err != nil {
		return nil, err
	}

	SCSIInquiry, err := dev.SCSIInquiry()
	if err != nil {
		return nil, err
	}

	// Check if device is an ATA device (For an ATA device VendorIdentication value should be equal to ATA    )
	if SCSIInquiry.VendorID == [8]byte{0x41, 0x54, 0x41, 0x20, 0x20, 0x20, 0x20, 0x20} {
		return &SATA{dev}, nil
	}

	return &dev, nil
}

// Open returns error if a SCSI device returns error when opened
func (d *SCSIDevice) Open() (err error) {
	d.fd, err = unix.Open(d.Name, unix.O_RDWR, 0600)
	return err
}

// Close returns error if a SCSI device is not closed
func (d *SCSIDevice) Close() error {
	return unix.Close(d.fd)
}

func (d *SCSIDevice) execSCSIGeneric(hdr *sgIOHeader) error {
	if err := ioctl.Ioctl(uintptr(d.fd), SGIO, uintptr(unsafe.Pointer(hdr))); err != nil {
		return err
	}

	// See http://www.t10.org/lists/2status.htm for SCSI status codes
	if hdr.info&SGInfoOkMask != SGInfoOk {
		err := sgIOErr{
			scsiStatus:   hdr.status,
			hostStatus:   hdr.hostStatus,
			driverStatus: hdr.driverStatus,
		}
		return err
	}

	return nil
}

// SCSIInquiry sends a SCSI INQUIRY command to a device and returns an InquiryResponse struct.
func (d *SCSIDevice) SCSIInquiry() (InquiryResponse, error) {
	var response InquiryResponse

	respBuf := make([]byte, INQRespLen)

	cdb := CDB6{SCSIInquiry}
	binary.BigEndian.PutUint16(cdb[3:], uint16(len(respBuf)))

	if err := d.sendCDB(cdb[:], &respBuf); err != nil {
		return response, err
	}

	binary.Read(bytes.NewBuffer(respBuf), utilities.NativeEndian, &response)

	return response, nil
}

// sendCDB sends a SCSI Command Descriptor Block to the device and writes the response into the
// supplied []byte pointer.
func (d *SCSIDevice) sendCDB(cdb []byte, respBuf *[]byte) error {
	senseBuf := make([]byte, 32)

	// Populate required fields of "sg_io_hdr_t" struct
	header := sgIOHeader{
		interfaceID:    'S',
		dxferDirection: SGDxferFromDev,
		timeout:        DefaultTimeout,
		cmdLen:         uint8(len(cdb)),
		mxSBLen:        uint8(len(senseBuf)),
		dxferLen:       uint32(len(*respBuf)),
		dxferp:         uintptr(unsafe.Pointer(&(*respBuf)[0])),
		cmdp:           uintptr(unsafe.Pointer(&cdb[0])),
		sbp:            uintptr(unsafe.Pointer(&senseBuf[0])),
	}

	return d.execSCSIGeneric(&header)
}

// modeSense sends a SCSI MODE SENSE(6) command to a device.
func (d *SCSIDevice) modeSense(pageNo, subPageNo, pageCtrl uint8) ([]byte, error) {
	respBuf := make([]byte, 64)

	cdb := CDB6{SCSIModeSense6}
	cdb[2] = (pageCtrl << 6) | (pageNo & 0x3f)
	cdb[3] = subPageNo
	cdb[4] = uint8(len(respBuf))

	if err := d.sendCDB(cdb[:], &respBuf); err != nil {
		return respBuf, err
	}

	return respBuf, nil
}

// readCapacity sends a SCSI READ CAPACITY(10) command to a device and returns the capacity in bytes.
func (d *SCSIDevice) readCapacity() (uint64, error) {
	respBuf := make([]byte, 8)
	cdb := CDB10{SCSIReadCapacity10}

	if err := d.sendCDB(cdb[:], &respBuf); err != nil {
		return 0, err
	}

	lastLBA := binary.BigEndian.Uint32(respBuf[0:]) // max. addressable LBA
	LBsize := binary.BigEndian.Uint32(respBuf[4:])  // logical block (i.e., sector) size
	capacity := (uint64(lastLBA) + 1) * uint64(LBsize)

	return capacity, nil
}

// PrintDiskInfo prints basic disk information
// Regular SCSI (including SAS, but excluding SATA)
func (d *SCSIDevice) PrintDiskInfo() error {
	capacity, _ := d.readCapacity()
	fmt.Printf("Capacity: %d bytes (%s)\n", capacity, utilities.ConvertBytes(capacity))

	// TODO : Fetch other disk attributes also such as serial no, vendor, etc
	// WIP
	response, _ := d.modeSense(RigidDiskDriveGeometryPage, 0, ModePageControlDefault)
	fmt.Printf("MODE SENSE buf: % x\n", response)

	respLen := response[0] + 1
	bdLen := response[3]
	offset := bdLen + 4
	fmt.Printf("respLen: %d, bdLen: %d, offset: %d\n",
		respLen, bdLen, offset)

	fmt.Printf("RPM: %d\n", binary.BigEndian.Uint16(response[offset+20:]))

	return nil
}

// GetDiskInfo returns smart disk info as well as basic disk info
func (d *SCSIDevice) GetDiskInfo() (DiskAttr, error) {
	capacity, _ := d.readCapacity()

	// TODO : Return all the basic disk attributes available for a particular disk
	DiskSmartAttr := DiskAttr{}
	DiskSmartAttr.UserCapacity = capacity

	return DiskSmartAttr, nil
}
