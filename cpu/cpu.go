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
		opcode := c.fetch()
		fmt.Printf("opcode: %#x\n%#v\n", opcode, c)

		err := c.decode(opcode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CPU) fetch() byte {
	value := c.mmu.ReadByte(c.pc)
	c.pc++

	return value
}

func (c *CPU) decode(opcode byte) error {
	switch opcode {
	case 0x00:
		c.nop()
	case 0x01, 0x11, 0x21:
		high, low := c.getRegisterPair(opcode)
		c.ld16(high, low)
	case 0x06, 0x0e, 0x16, 0x1e, 0x26, 0x2e, 0x3e:
		pointer := c.getDestRegister(opcode)
		c.ld8(pointer)
	case 0x32:
		c.ldhld()
	case 0xa8, 0xa9, 0xaa, 0xab, 0xac, 0xad, 0xaf:
		pointer := c.getSourceRegister(opcode)
		c.xor(*pointer)
	case 0xc3:
		c.jp()
	default:
		return &UnknownOpcodeError{opcode: opcode}
	}

	c.cycles += cycles[opcode]

	return nil
}

func (c *CPU) getRegister(code byte) *byte {
	switch code {
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

func (c *CPU) getDestRegister(opcode byte) *byte {
	code := opcode >> 3 & 0x7
	return c.getRegister(code)
}

func (c *CPU) getSourceRegister(opcode byte) *byte {
	code := opcode & 0x7
	return c.getRegister(code)
}

func (c *CPU) getRegisterPair(opcode byte) (*byte, *byte) {
	switch opcode >> 4 & 0x3 {
	case 0:
		return &c.b, &c.c
	case 1:
		return &c.d, &c.e
	case 2:
		return &c.h, &c.l
	}

	return nil, nil
}

func (c *CPU) setFlags(value byte) {
	c.f |= value
}

func (c *CPU) resetFlags(value byte) {
	c.f &^= value
}

func (c *CPU) nop() {

}

func (c *CPU) ld16(high, low *byte) {
	*low = c.fetch()
	*high = c.fetch()
}

func (c *CPU) ld8(register *byte) {
	*register = c.fetch()
}

func (c *CPU) ldhld() {
	address := uint16(c.h)<<8 | uint16(c.l)
	c.mmu.WriteByte(address, c.a)

	address--
	c.h = byte(address >> 8)
	c.l = byte(address & 0xff)
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
