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

// Package smartinfo is a pure Go SMART library.
//
package smartinfo

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/openebs/smart/ioctl"
	"github.com/openebs/smart/scsismart"
)

// ScanDevices discover and return the list of scsi devices.
func ScanDevices() []scsismart.SCSIDevice {
	var devices []scsismart.SCSIDevice

	// Find all SCSI disk devices
	files, err := filepath.Glob("/dev/sd*[^0-9]")
	if err != nil {
		return devices
	}

	for _, file := range files {
		devices = append(devices, scsismart.SCSIDevice{Name: file})
	}

	return devices
}

// Scan prints the list of SCSI devices
func Scan() {
	for _, device := range ScanDevices() {
		fmt.Printf("%#v\n", device)
	}

}

// DiskDetail returs details(disk attributes and their values such as vendor,serialno,etc) of a disk
func DiskDetail(device string) scsismart.DiskAttr {
	fmt.Println("Openebs Smart GO Implementation")
	fmt.Printf("Built with %s on %s (%s)\n\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	ioctl.CapabilitiesCheck()

	var diskDetails scsismart.DiskAttr

	if device != "" {
		var (
			d   scsismart.Dev // interface
			err error
		)

		d, err = scsismart.DetectSCSIType(device)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer d.Close()

		diskDetails, err = d.GetDiskInfo()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	}
	return diskDetails
}
