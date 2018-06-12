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

// Functions for SCSI-ATA Translation.

package scsismart

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/openebs/smart/atasmart"
	"github.com/openebs/smart/utilities"
)

// SATA is a simple wrapper around an embedded SCSIDevice type, which handles sending ATA
// commands via SCSI pass-through (SCSI-ATA Translation).
type SATA struct {
	SCSIDevice
}

// AtaIdentify sends SCSI_ATA_PASSTHRU_16 command and read data from the response based on the defined ATA IDENTIFY STRUCT in ataidentify.go
func (d *SATA) AtaIdentify() (atasmart.IdentDevData, error) {
	var identifyBuf atasmart.IdentDevData

	responseBuf := make([]byte, 512)

	cdb16 := CDB16{SCSIATAPassThru16}
	cdb16[1] = 0x08 // ATA protocol (4 << 1, PIO data-in)
	cdb16[2] = 0x0e // BYT_BLOK = 1, T_LENGTH = 2, T_DIR = 1
	cdb16[14] = atasmart.AtaIdentifyDevice

	if err := d.sendCDB(cdb16[:], &responseBuf); err != nil {
		return identifyBuf, fmt.Errorf("sendCDB ATA IDENTIFY: %v", err)
	}

	binary.Read(bytes.NewBuffer(responseBuf), utilities.NativeEndian, &identifyBuf)

	return identifyBuf, nil
}

// GetDiskInfo returns all the disk attributes and smart info for a particular SATA device
func (d *SATA) GetDiskInfo() (DiskAttr, error) {
	// Standard SCSI INQUIRY command
	inqResp, err := d.SCSIInquiry()
	if err != nil {
		return DiskAttr{}, fmt.Errorf("SgExecute INQUIRY: %v", err)
	}

	// inqCapacity is the total capacity of a disk in bytes
	inqCapacity, err := d.readCapacity()
	if err != nil {
		return DiskAttr{}, fmt.Errorf("SgExecute readCapacity: %v", err)
	}

	identifyBuf, err := d.AtaIdentify()
	if err != nil {
		return DiskAttr{}, err
	}

	LogicalSec, PhysicalSec := identifyBuf.GetSectorSize()

	SATASmartAttr := DiskAttr{}
	SATASmartAttr.SCSIInquiry = inqResp
	SATASmartAttr.UserCapacity = inqCapacity
	SATASmartAttr.LBSize = LogicalSec
	SATASmartAttr.PBSize = PhysicalSec
	SATASmartAttr.SerialNumber = string(identifyBuf.GetSerialNumber())
	SATASmartAttr.LuWWNDeviceID = identifyBuf.GetWWN()
	SATASmartAttr.FirmwareRevision = string(identifyBuf.GetFirmwareRevision())
	SATASmartAttr.ModelNumber = string(identifyBuf.GetModelNumber())
	SATASmartAttr.RotationRate = identifyBuf.RotationRate
	SATASmartAttr.ATAMajorVersion = identifyBuf.GetATAMajorVersion()
	SATASmartAttr.ATAMinorVersion = identifyBuf.GetATAMinorVersion()
	SATASmartAttr.Transport = identifyBuf.Transport()

	return SATASmartAttr, nil
}

// PrintDiskInfo prints all the available information for a SATA disk (both basic attr and smart attr)
func (d *SATA) PrintDiskInfo() error {
	// Standard SCSI INQUIRY command
	inqResp, err := d.SCSIInquiry()
	if err != nil {
		return fmt.Errorf("SgExecute INQUIRY: %v", err)
	}

	fmt.Println("SCSI INQUIRY:", inqResp)

	// inqCapacity is the total capacity of a disk in bytes
	inqCapacity, err := d.readCapacity()
	if err != nil {
		return fmt.Errorf("SgExecute readCapacity: %v", err)
	}

	fmt.Printf("User Capacity:%v bytes (%v)\n", inqCapacity, utilities.ConvertBytes(inqCapacity))

	identifyBuf, err := d.AtaIdentify()
	if err != nil {
		return err
	}

	LogicalSec, PhysicalSec := identifyBuf.GetSectorSize()

	fmt.Println("\nATA IDENTIFY data :")
	fmt.Printf("Serial Number: %s\n", identifyBuf.GetSerialNumber())
	fmt.Printf("Model Number: %s\n", identifyBuf.GetModelNumber())
	fmt.Println("LU WWN Device Id:", identifyBuf.GetWWN())
	fmt.Printf("Firmware Revision: %s\n", identifyBuf.GetFirmwareRevision())
	fmt.Println("ATA Major Version:", identifyBuf.GetATAMajorVersion())
	fmt.Println("ATA Minor Version:", identifyBuf.GetATAMinorVersion())
	fmt.Printf("Sector Size: %d bytes logical, %d bytes physical\n", LogicalSec, PhysicalSec)
	fmt.Printf("Rotation Rate: %d\n", identifyBuf.RotationRate)
	fmt.Printf("SMART support available: %v\n", identifyBuf.Word87>>14 == 1)
	fmt.Printf("SMART support enabled: %v\n", identifyBuf.Word85&0x1 != 0)
	fmt.Println("Transport:", identifyBuf.Transport())

	return nil
}
