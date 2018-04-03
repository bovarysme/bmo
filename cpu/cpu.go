package cpu

import (
	"fmt"

	"github.com/bovarysme/bmo/mmu"
)

// CPU flags' masks
const (
	carry byte = 1 << (4 + iota)
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

	// Stack pointer
	sp uint16
	// Program counter
	pc uint16

	// Interrupt Master Enable flag
	ime bool

	cycles int

	mmu *mmu.MMU
}

func NewCPU(mmu *mmu.MMU) *CPU {
	return &CPU{
		a: 0x01,
		f: 0xb0,
		b: 0x00,
		c: 0x13,
		d: 0x00,
		e: 0xd8,
		h: 0x01,
		l: 0x4d,

		sp: 0xfffe,
		pc: 0x0100,

		mmu: mmu,
	}
}

func (c *CPU) Step() error {
	opcode := c.fetch()
	fmt.Printf("opcode: %#x\n%#v\n\n", opcode, c)

	err := c.decode(opcode)

	return err
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
	case 0x04, 0x0c, 0x14, 0x1c, 0x24, 0x2c, 0x34:
		pointer := c.getDestRegister(opcode)
		c.inc(pointer)
	case 0x05, 0x0d, 0x15, 0x1d, 0x25, 0x2d, 0x3d:
		pointer := c.getDestRegister(opcode)
		c.dec(pointer)
	case 0x06, 0x0e, 0x16, 0x1e, 0x26, 0x2e, 0x3e:
		pointer := c.getDestRegister(opcode)
		c.ld8(pointer)
	case 0x20, 0x28, 0x30, 0x38:
		condition := c.getCondition(opcode)
		c.jr(condition)
	case 0x32:
		c.ldd()
	case 0x37:
		c.scf()
	case 0x3f:
		c.ccf()
	case 0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa7, 0xe6:
		value := c.getArithmeticValue(opcode)
		c.and(value)
	case 0xa8, 0xa9, 0xaa, 0xab, 0xac, 0xad, 0xaf, 0xee:
		value := c.getArithmeticValue(opcode)
		c.xor(value)
	case 0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5, 0xb7, 0xf6:
		value := c.getArithmeticValue(opcode)
		c.or(value)
	case 0xb8, 0xb9, 0xba, 0xbb, 0xbc, 0xbd, 0xbf, 0xfe:
		value := c.getArithmeticValue(opcode)
		c.cp(value)
	case 0xc3:
		c.jp()
	case 0xe0:
		c.sth()
	case 0xf0:
		c.ldh()
	case 0xf3:
		c.di()
	case 0xfb:
		c.ei()
	default:
		return &UnknownOpcodeError{opcode: opcode}
	}

	c.cycles += cycles[opcode]

	return nil
}

func (c *CPU) getCondition(opcode byte) bool {
	switch opcode >> 3 & 0x3 {
	case 0:
		return c.f&zero == 0
	case 1:
		return c.f&zero == zero
	case 2:
		return c.f&carry == 0
	case 3:
		return c.f&carry == carry
	}

	return false
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

func (c *CPU) getArithmeticValue(opcode byte) byte {
	// If the instruction is arithmetic and has a d8 operand
	if opcode&0xc6 == 0xc6 && opcode&1 == 0 {
		return c.fetch()
	} else {
		return *c.getSourceRegister(opcode)
	}
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

func (c *CPU) inc(register *byte) {
	*register++

	c.resetFlags(zero | substract | halfCarry)

	if *register == 0 {
		c.setFlags(zero | halfCarry)
	}
}

func (c *CPU) dec(register *byte) {
	*register--

	c.resetFlags(zero | halfCarry)
	c.setFlags(substract)

	if *register == 0 {
		c.setFlags(zero)
	} else if *register == 0xff {
		c.setFlags(halfCarry)
	}
}

func (c *CPU) ld8(register *byte) {
	*register = c.fetch()
}

func (c *CPU) jr(condition bool) {
	// XXX: [-128; 127] when the docs say [-127; 129]
	steps := int8(c.fetch())

	if condition {
		c.cycles++

		// XXX: is there a cleaner way to do this?
		c.pc = uint16(int16(c.pc) + int16(steps))
	}
}

func (c *CPU) ldd() {
	address := uint16(c.h)<<8 | uint16(c.l)
	c.mmu.WriteByte(address, c.a)

	address--
	c.h = byte(address >> 8)
	c.l = byte(address & 0xff)
}

// Sets the carry flag.
func (c *CPU) scf() {
	c.resetFlags(substract | halfCarry)
	c.setFlags(carry)
}

// Flips the carry flag.
func (c *CPU) ccf() {
	c.resetFlags(substract | halfCarry)
	c.f ^= carry
}

func (c *CPU) and(value byte) {
	c.a &= value

	c.resetFlags(zero | substract | carry)
	c.setFlags(halfCarry)

	if c.a == 0 {
		c.setFlags(zero)
	}
}

func (c *CPU) xor(value byte) {
	c.a ^= value

	c.resetFlags(zero | substract | halfCarry | carry)
	if c.a == 0 {
		c.setFlags(zero)
	}
}

func (c *CPU) or(value byte) {
	c.a |= value

	c.resetFlags(zero | substract | halfCarry | carry)
	if c.a == 0 {
		c.setFlags(zero)
	}
}

func (c *CPU) cp(value byte) {
	c.resetFlags(zero | halfCarry | carry)
	c.setFlags(substract)

	if value < c.a {
		c.setFlags(halfCarry)
	} else if value == c.a {
		c.setFlags(zero)
	} else {
		c.setFlags(carry)
	}
}

func (c *CPU) jp() {
	c.pc = c.mmu.ReadWord(c.pc)
}

func (c *CPU) sth() {
	address := 0xff00 + uint16(c.fetch())
	c.mmu.WriteByte(address, c.a)
}

func (c *CPU) ldh() {
	address := 0xff00 + uint16(c.fetch())
	c.a = c.mmu.ReadByte(address)
}

func (c *CPU) di() {
	c.ime = false
}

func (c *CPU) ei() {
	c.ime = true
}
