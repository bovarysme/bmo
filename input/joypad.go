package input

import (
	"github.com/bovarysme/bmo/interrupt"
)

const P1 uint16 = 0xff00

const (
	Right byte = iota
	Left
	Up
	Down
	A
	B
	Select
	Start
)

const (
	selectDirectionKeys byte = 1 << (4 + iota)
	selectButtonKeys
)

type Joypad struct {
	ic *interrupt.IC

	p1 byte

	directionKeysState byte
	buttonKeysState    byte
}

func NewJoypad(ic *interrupt.IC) *Joypad {
	return &Joypad{
		ic: ic,

		directionKeysState: 0xf,
		buttonKeysState:    0xf,
	}
}

func (j *Joypad) ReadByte(address uint16) byte {
	selected := j.getSelected()
	if selected != nil {
		j.p1 |= *selected
	}

	return j.p1
}

func (j *Joypad) WriteByte(address uint16, value byte) {
	j.p1 = value & 0x30
}

func (j *Joypad) SetKey(mask byte) {
	state := j.getState(mask)
	*state &^= 1 << (mask % 4)
}

func (j *Joypad) ResetKey(mask byte) {
	state := j.getState(mask)
	*state |= 1 << (mask % 4)

	j.ic.Request(interrupt.Joypad)
}

func (j *Joypad) getSelected() *byte {
	var selected *byte

	if j.p1&selectDirectionKeys == 0 {
		selected = &j.directionKeysState
	} else if j.p1&selectButtonKeys == 0 {
		selected = &j.buttonKeysState
	}

	return selected
}

func (j *Joypad) getState(mask byte) *byte {
	var state *byte

	if mask <= 3 {
		state = &j.directionKeysState
	} else {
		state = &j.buttonKeysState
	}

	return state
}
