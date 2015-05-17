/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"errors"
	"fmt"
	"github.com/riobard/go-virtualbox"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var cache string = "/var/tmp/.utahcache"

func CopyFile(src, dest string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, input)

	return err
}

func DownloadToCache(url, filename string) error {
	dest := filepath.Join(cache, filename)

	err := os.MkdirAll(cache, 0755)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Println("Download of", url, "failed", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Received a return code of ", resp.StatusCode))
	}

	if _, err := os.Stat(dest); err == nil || err.(*os.PathError).Err != syscall.ENOENT {
		if err == nil {
			log.Println("cached file", filename, "already exists")
		}
		return err
	}
	output, err := ioutil.TempFile(cache, "utahtmp")
	if err != nil {
		log.Println("Getting a temporary file failed", err)
		return err
	}
	_, err = io.Copy(output, resp.Body)
	if err != nil {
		log.Println("cache, o.Name, dest", cache, output.Name(), dest)
		return err
	}
	err = os.Rename(output.Name(), dest)
	return err
}

func main() {
	virtualbox.Verbose = true
	log.SetFlags(log.Lshortfile)

	err := DownloadToCache("https://cloud-images.ubuntu.com/trusty/current/trusty-server-cloudimg-amd64-disk1.img", "trusty-cloud.img")
	if err != nil {
		fmt.Println("Failed to populate the image cache", err)
		return
	}

	_, err = virtualbox.CreateMachine("utah-test0", "")
	if err != nil {
		fmt.Println("create machine failed", err)
		return
	}
	machines, _ := virtualbox.ListMachines()
	fmt.Println(machines)

	utah := machines[0]

	// set up storage
	// Chose things that looked like VB defaults, and magic numbers from boot2docker-cli
	err = utah.AddStorageCtl("defaultctlr", virtualbox.StorageController{SysBus: virtualbox.SysBusSATA, Ports: 4, Chipset: virtualbox.CtrlIntelAHCI, HostIOCache: true, Bootable: true})
	if err != nil {
		fmt.Println("adding storage controller failed", err)
		return
	}

	wd, _ := os.Getwd()
	fmt.Println(wd)
	err = CopyFile("trusty-cloud.vdi", "temp.vdi")
	if err != nil {
		fmt.Println("creating a backing store failed", err)
		return
	}

	err = utah.AttachStorage("defaultctlr", virtualbox.StorageMedium{Port: 1, Device: 0, DriveType: virtualbox.DriveHDD, Medium: filepath.Join(wd, "temp.vdi")})
	if err != nil {
		fmt.Println("attaching storage failed", err)
		return
	}

	// set up network
	nic := virtualbox.NIC{virtualbox.NICNetHostonly, virtualbox.VirtIO, "vboxnet0"}
	err = utah.SetNIC(1, nic)
	if err != nil {
		fmt.Println("nic setup failed", err)
		return
	}

	// boot
	log.Println("vb state", utah.State)
	utah.Refresh()
	err = utah.Start()
	if err != nil {
		fmt.Println("starting the machine failed", err)
		return
	}
	log.Println("vb state", utah.State)
	time.Sleep(10)
	utah.Refresh()
	log.Println("vb state", utah.State)

	// poweroff
	err = utah.Poweroff()
	if err != nil {
		fmt.Println("powering off the machine failed", err)
		return
	}
	utah.Refresh()
	log.Println("powered off the machine. State:", utah.State)

	// delete
	err = utah.Delete()
	if err != nil {
		fmt.Println("deleting the machine failed", err)
		return
	}
}
