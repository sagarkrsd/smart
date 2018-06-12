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

// OpenEBS smartinfo go library for disks.

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/openebs/smart/ioctl"
	"github.com/openebs/smart/scsismart"
	"github.com/openebs/smart/smartinfo"
)

func scanDevices() {
	for _, device := range smartinfo.ScanDevices() {
		fmt.Printf("%#v\n", device)
	}

}

func main() {
	fmt.Println("OpenEBS smart go library")
	fmt.Printf("Built with %s on %s (%s)\n\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	devPath := flag.String("devPath", "", "SATA device path from which to read SMART attributes, e.g., /dev/sda")
	devScan := flag.Bool("devScan", false, "scan for devices that support smart")
	flag.Parse()

	// check if required permissions are set or not
	ioctl.CapabilitiesCheck()

	if *devPath != "" {
		var (
			d   scsismart.Dev // interface
			err error
		)

		d, err = scsismart.DetectSCSIType(*devPath)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer d.Close()

		if err := d.PrintDiskInfo(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else if *devScan {
		scanDevices()
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}
}
