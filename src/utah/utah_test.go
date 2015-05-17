package utah_test

import	(
		"utah"
		"testing"
		"github.com/riobard/go-virtualbox"
		"fmt"
		"time"
		"log"
		"path/filepath"
		)

func TestLifecycle(t *testing.T) {
	virtualbox.Verbose = true
	log.SetFlags(log.Lshortfile)

	// Ideally, I'd use something smaller
	coreos_vm := "coreos-current-prod.vmdk"
	coreos_vd := "coreos-current-prod.vdi"
	err := utah.DownloadToCache("http://stable.release.core-os.net/amd64-usr/current/coreos_production_virtualbox_image.vmdk.bz2", coreos_vm)
	if err != nil {
		fmt.Println("Failed to populate the image cache", err)
		return
	}

	err = utah.ConvertToVDI(filepath.Join(utah.Cache, coreos_vm), coreos_vd)
	if err != nil {
		fmt.Println("Conversion failed", err)
		return
	}

	new_machine, err := utah.CreateMachine("utah-test0", coreos_vd)
	if err != nil {
		fmt.Println("Creating a machine failed", err)
	}

	// boot
	log.Println("vb state", new_machine.State())
	err = new_machine.Start()
	if err != nil {
		fmt.Println("starting the machine failed", err)
		return
	}
	log.Println("vb state", new_machine.State())
	time.Sleep(10)
	log.Println("vb state", new_machine.State())

	// poweroff
	err = new_machine.Stop()
	if err != nil {
		fmt.Println("powering off the machine failed", err)
		return
	}
	log.Println("powered off the machine. State:", new_machine.State())

	// delete
	err = new_machine.Delete()
	if err != nil {
		fmt.Println("deleting the machine failed", err)
		return
	}
}