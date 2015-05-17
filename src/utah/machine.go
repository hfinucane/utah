/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package utah

import (
	"github.com/riobard/go-virtualbox"
)

type Machine struct {
	vb_machine  *virtualbox.Machine
	name, image string
}

type MachineState string

const (
	Poweroff = MachineState("poweroff")
	Running  = MachineState("running")
	Paused   = MachineState("paused")
	Saved    = MachineState("saved")
	Aborted  = MachineState("aborted")
	Missing  = MachineState("missing")
)

func (this *Machine) Start() error {
	err := this.vb_machine.Refresh()
	if err != nil {
		return err
	}
	err = this.vb_machine.Start()
	return err
}

func (this *Machine) Stop() error {
	return this.vb_machine.Poweroff()
}

func (this *Machine) Delete() error {
	return this.vb_machine.Delete()
}

func (this *Machine) State() MachineState {
	err := this.vb_machine.Refresh()
	if err != nil {
		return Missing
	}
	return MachineState(this.vb_machine.State)
}
