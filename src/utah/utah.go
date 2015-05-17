/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package utah

import (
	"errors"
	"fmt"
	"github.com/riobard/go-virtualbox"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

var Cache string = "/var/tmp/.utahcache"

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
	dest := filepath.Join(Cache, filename)

	err := os.MkdirAll(Cache, 0755)
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
	output, err := ioutil.TempFile(Cache, "utahtmp")
	if err != nil {
		log.Println("Getting a temporary file failed", err)
		return err
	}
	_, err = io.Copy(output, resp.Body)
	if err != nil {
		log.Println("cache, o.Name, dest", Cache, output.Name(), dest)
		return err
	}
	err = os.Rename(output.Name(), dest)
	return err
}

func ConvertToVDI(src, dest string) error {
	if _, err := os.Stat(dest); err == nil || err.(*os.PathError).Err != syscall.ENOENT {
		if err == nil {
			log.Println(dest, " already exists")
		}
		return err
	}

	cmd := exec.Command("qemu-img", "convert", "-O", "vdi", src, dest)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Run()
	switch err.(type) {
	case *exec.ExitError:
		stderr_bytes, err := ioutil.ReadAll(stderr)
		if err != nil {
			return err
		}
		return errors.New(string(stderr_bytes))
	default:
		return err
	}
}

func CreateMachine(name, image string) (*Machine, error) {
	_, err := virtualbox.CreateMachine(name, "")
	if err != nil {
		fmt.Println("create machine failed", err)
		return nil, err
	}
	machines, _ := virtualbox.ListMachines()
	fmt.Println(machines)

	// XXX TODO FIXME actually get the right machine out of here
	new_machine := machines[len(machines)-1]

	// set up storage
	// Chose things that looked like VB defaults, and magic numbers from boot2docker-cli
	err = new_machine.AddStorageCtl("defaultctlr", virtualbox.StorageController{SysBus: virtualbox.SysBusSATA, Ports: 4, Chipset: virtualbox.CtrlIntelAHCI, HostIOCache: true, Bootable: true})
	if err != nil {
		fmt.Println("adding storage controller failed", err)
		return nil, err
	}

	wd, _ := os.Getwd()
	fmt.Println(wd)
	err = CopyFile(image, "temp.vdi")
	if err != nil {
		fmt.Println("creating a backing store failed", err)
		return nil, err
	}

	err = new_machine.AttachStorage("defaultctlr", virtualbox.StorageMedium{Port: 1, Device: 0, DriveType: virtualbox.DriveHDD, Medium: filepath.Join(wd, "temp.vdi")})
	if err != nil {
		fmt.Println("attaching storage failed", err)
		return nil, err
	}

	// set up network
	nic := virtualbox.NIC{virtualbox.NICNetHostonly, virtualbox.VirtIO, "vboxnet0"}
	err = new_machine.SetNIC(1, nic)
	if err != nil {
		fmt.Println("nic setup failed", err)
		return nil, err
	}

	return &Machine{new_machine, name, image}, nil
}
