package cpu

import (
	"fmt"

	"github.com/bovarysme/bmo/mmu"
)

const (
	carry byte = 1 << (iota + 4)
	halfCarry
	substract
	zero
)

var cycles = [...]int{
	1, 3, 2, 2, 1, 1, 2, 1, 5, 2, 2, 2, 1, 1, 2, 1, // 0x0
	1, 3, 2, 2, 1, 1, 2, 1, 3, 2, 2, 2, 1, 1, 2, 1, // 0x1
	2, 3, 2, 2, 1, 1, 2, 1, 2, 2, 2, 2, 1, 1, 2, 1, // 0x2
	2, 3, 2, 2, 3, 3, 3, 1, 2, 2, 2, 2, 1, 1, 2, 1, // 0x3
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 0x4
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 0x5
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 0x6
	2, 2, 2, 2, 2, 2, 1, 2, 1, 1, 1, 1, 1, 1, 2, 1, // 0x7
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 0x8
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 0x9
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 0xa
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 0xb
	2, 3, 3, 4, 3, 4, 2, 4, 2, 4, 3, 1, 3, 6, 2, 4, // 0xc
	2, 3, 3, 0, 3, 4, 2, 4, 2, 4, 3, 0, 3, 0, 2, 4, // 0xd
	3, 3, 2, 0, 0, 4, 2, 4, 4, 1, 4, 0, 0, 0, 2, 4, // 0xe
	3, 3, 2, 1, 0, 4, 2, 4, 3, 2, 4, 1, 0, 0, 2, 4, // 0xf
}

type UnknownOpcodeError struct {
	opcode byte
}

func (e UnknownOpcodeError) Error() string {
	return fmt.Sprintf("Unknown opcode: %#x", e.opcode)
}

type CPU struct {
	// Registers
	a byte
	f byte
	b byte
	c byte
	d byte
	e byte
	h byte
	l byte

	pc     uint16
	cycles int

	mmu *mmu.MMU
}

func NewCPU(mmu *mmu.MMU) *CPU {
	return &CPU{
		pc: 0x100,

		mmu: mmu,
	}
}

func (c *CPU) Run() error {
	for {
		opcode := c.mmu.ReadByte(c.pc)
		c.pc++

		err := c.Decode(opcode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CPU) Decode(opcode byte) error {
	switch opcode {
	case 0x00:
		c.nop()
	case 0xa8, 0xa9, 0xaa, 0xab, 0xac, 0xad, 0xaf:
		pointer := c.getRegister(opcode)
		c.xor(*pointer)
	case 0xc3:
		c.jp()
	default:
		return &UnknownOpcodeError{opcode: opcode}
	}

	c.cycles += cycles[opcode]

	return nil
}

func (c *CPU) getRegister(opcode byte) *byte {
	switch opcode & 0x7 {
	case 0:
		return &c.b
	case 1:
		return &c.c
	case 2:
		return &c.d
	case 3:
		return &c.e
	case 4:
		return &c.h
	case 5:
		return &c.l
	case 7:
		return &c.a
	}

	return nil
}

func (c *CPU) setFlags(value byte) {
	c.f |= value
}

func (c *CPU) resetFlags(value byte) {
	c.f &^= value
}

func (c *CPU) nop() {

}

func (c *CPU) xor(value byte) {
	c.a ^= value

	// XXX: modifying the zero flag twice
	c.resetFlags(zero | substract | halfCarry | carry)
	if c.a == 0 {
		c.setFlags(zero)
	}
}

func (c *CPU) jp() {
	c.pc = c.mmu.ReadWord(c.pc)
}
