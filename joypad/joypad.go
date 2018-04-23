package joypad

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
	directionSelect byte = 1 << (iota + 4)
	buttonSelect
)

type Joypad struct {
	ic *interrupt.IC

	p1 byte

	direction byte
	button    byte
}

func NewJoypad(ic *interrupt.IC) *Joypad {
	return &Joypad{
		ic: ic,

		direction: 0xf,
		button:    0xf,
	}
}

func (j *Joypad) SetKey(mask byte) {
	state := j.getState(mask)
	*state &^= 1 << (mask % 4)
}

func (j *Joypad) ResetKey(mask byte) {
	state := j.getState(mask)
	*state |= 1 << (mask % 4)
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

func (j *Joypad) getSelected() *byte {
	var selected *byte

	if j.p1&directionSelect == 0 {
		selected = &j.direction
	} else if j.p1&buttonSelect == 0 {
		selected = &j.button
	}

	return selected
}

func (j *Joypad) getState(mask byte) *byte {
	var state *byte

	if mask <= 3 {
		state = &j.direction
	} else {
		state = &j.button
	}

	return state
}
