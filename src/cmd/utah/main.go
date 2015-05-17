/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"fmt"
	"github.com/riobard/go-virtualbox"
	"log"
	"path/filepath"
	"time"
	"utah"
)

func main() {
	virtualbox.Verbose = true
	log.SetFlags(log.Lshortfile)

	err := utah.DownloadToCache("https://cloud-images.ubuntu.com/trusty/current/trusty-server-cloudimg-amd64-disk1.img", "trusty-cloud.img")
	if err != nil {
		fmt.Println("Failed to populate the image cache", err)
		return
	}

	err = utah.ConvertToVDI(filepath.Join(utah.Cache, "trusty-cloud.img"), "trusty-cloud.vdi")
	if err != nil {
		fmt.Println("Conversion failed", err)
		return
	}

	new_machine, err := utah.CreateMachine("utah-test0", "trusty-cloud.vdi")
	if err != nil {
		fmt.Println("Creating a machine failed", err)
	}

	// boot
	log.Println("vb state", new_machine.State)
	new_machine.Refresh()
	err = new_machine.Start()
	if err != nil {
		fmt.Println("starting the machine failed", err)
		return
	}
	log.Println("vb state", new_machine.State)
	time.Sleep(10)
	new_machine.Refresh()
	log.Println("vb state", new_machine.State)

	// poweroff
	err = new_machine.Poweroff()
	if err != nil {
		fmt.Println("powering off the machine failed", err)
		return
	}
	new_machine.Refresh()
	log.Println("powered off the machine. State:", new_machine.State)

	// delete
	err = new_machine.Delete()
	if err != nil {
		fmt.Println("deleting the machine failed", err)
		return
	}
}
