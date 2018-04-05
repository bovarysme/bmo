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

type UnknownPrefixOpcodeError struct {
	opcode byte
}

func (e UnknownPrefixOpcodeError) Error() string {
	return fmt.Sprintf("Unknown prefix opcode: %#x", e.opcode)
}

type Operand interface {
	Get() byte
	Set(value byte)
}

type RegisterOperand struct {
	register *byte
}

func (o *RegisterOperand) Get() byte {
	return *o.register
}

func (o *RegisterOperand) Set(value byte) {
	*o.register = value
}

type MemoryOperand struct {
	address uint16
	mmu     *mmu.MMU
}

func (o *MemoryOperand) Get() byte {
	return o.mmu.ReadByte(o.address)
}

func (o *MemoryOperand) Set(value byte) {
	o.mmu.WriteByte(o.address, value)
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

func (c *CPU) fetchWord() uint16 {
	value := c.mmu.ReadWord(c.pc)
	c.pc += 2

	return value
}

func (c *CPU) decode(opcode byte) error {
	switch opcode {
	case 0x00:
		c.nop()
	case 0x01, 0x11, 0x21:
		high, low := c.decodeRegisterPair(opcode)
		c.ld16(high, low)
	case 0x04, 0x0c, 0x14, 0x1c, 0x24, 0x2c, 0x34, 0x3c:
		operand := c.getDestOperand(opcode)
		c.inc(operand)
	case 0x05, 0x0d, 0x15, 0x1d, 0x25, 0x2d, 0x35, 0x3d:
		operand := c.getDestOperand(opcode)
		c.dec(operand)
	case 0x06, 0x0e, 0x16, 0x1e, 0x26, 0x2e, 0x36, 0x3e,
		0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47,
		0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f,
		0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57,
		0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f,
		0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67,
		0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f,
		0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x77,
		0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f:

		operand := c.getDestOperand(opcode)
		value := c.getSourceValue(opcode)
		c.ld(operand, value)
	case 0x0b, 0x1b, 0x2b:
		high, low := c.decodeRegisterPair(opcode)
		c.dec16(high, low)
	case 0x20, 0x28, 0x30, 0x38:
		condition := c.decodeCondition(opcode)
		c.jr(condition)
	case 0x2a:
		c.ldi()
	case 0x2f:
		c.cpl()
	// TODO: merge with ld16?
	case 0x31:
		c.ldsp16()
	case 0x32:
		c.std()
	case 0x37:
		c.scf()
	case 0x3f:
		c.ccf()
	case 0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xe6:
		value := c.getSourceValue(opcode)
		c.and(value)
	case 0xa8, 0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xee:
		value := c.getSourceValue(opcode)
		c.xor(value)
	case 0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xf6:
		value := c.getSourceValue(opcode)
		c.or(value)
	case 0xb8, 0xb9, 0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf, 0xfe:
		value := c.getSourceValue(opcode)
		c.cp(value)
	case 0xc3:
		c.jp()
	case 0xc9:
		c.ret()
	case 0xcb:
		err := c.decodePrefix()
		if err != nil {
			return err
		}
	case 0xcd:
		c.call()
	case 0xe0:
		c.sta8()
	case 0xe2:
		c.stc()
	case 0xea:
		c.sta16()
	case 0xf0:
		c.lda8()
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

func (c *CPU) decodePrefix() error {
	opcode := c.fetch()

	switch {
	case opcode >= 0x40 && opcode <= 0x7f:
		bit := c.decodeBit(opcode)
		operand := c.getSourceOperand(opcode)
		c.bit(bit, operand)
	case opcode >= 0x80 && opcode <= 0xbf:
		bit := c.decodeBit(opcode)
		operand := c.getSourceOperand(opcode)
		c.res(bit, operand)
	case opcode >= 0xc0 && opcode <= 0xff:
		bit := c.decodeBit(opcode)
		operand := c.getSourceOperand(opcode)
		c.set(bit, operand)
	default:
		return &UnknownPrefixOpcodeError{opcode: opcode}
	}

	return nil
}

func (c *CPU) popStack() uint16 {
	value := c.mmu.ReadWord(c.sp)
	c.sp += 2

	return value
}

func (c *CPU) pushStack(value uint16) {
	c.mmu.WriteByte(c.sp-1, byte(c.pc>>8))
	c.mmu.WriteByte(c.sp-2, byte(c.pc&0xff))
	c.sp -= 2
}

func (c *CPU) decodeBit(opcode byte) byte {
	return opcode >> 3 & 0x3
}

func (c *CPU) decodeCondition(opcode byte) bool {
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

func (c *CPU) decodeRegister(code byte) *byte {
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

func (c *CPU) decodeDestRegister(opcode byte) *byte {
	code := opcode >> 3 & 0x7
	return c.decodeRegister(code)
}

func (c *CPU) decodeSourceRegister(opcode byte) *byte {
	code := opcode & 0x7
	return c.decodeRegister(code)
}

func (c *CPU) decodeRegisterPair(opcode byte) (*byte, *byte) {
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

func (c *CPU) getAddress() uint16 {
	return uint16(c.h)<<8 | uint16(c.l)
}

func (c *CPU) getOperand(register *byte) Operand {
	if register != nil {
		return &RegisterOperand{register: register}
	}

	address := c.getAddress()
	return &MemoryOperand{
		address: address,
		mmu:     c.mmu,
	}
}

func (c *CPU) getSourceOperand(opcode byte) Operand {
	register := c.decodeSourceRegister(opcode)
	return c.getOperand(register)
}

func (c *CPU) getDestOperand(opcode byte) Operand {
	register := c.decodeDestRegister(opcode)
	return c.getOperand(register)
}

func (c *CPU) getSourceValue(opcode byte) byte {
	// If the instruction has a register source operand.
	register := c.decodeSourceRegister(opcode)
	if register != nil {
		return *register
	}

	// If the instruction has a d8 source operand (i.e. its 2 highest bits are
	// either 0b00 or 0b11).
	if opcode>>6 == 0 || opcode>>6 == 0x3 {
		return c.fetch()
	}

	// Else the instruction has a (HL) source operand.
	address := c.getAddress()
	return c.mmu.ReadByte(address)
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

func (c *CPU) ldsp16() {
	c.sp = c.fetchWord()
}

func (c *CPU) inc(operand Operand) {
	value := operand.Get() + 1
	operand.Set(value)

	c.resetFlags(zero | substract | halfCarry)

	if value == 0 {
		c.setFlags(zero | halfCarry)
	}
}

func (c *CPU) dec(operand Operand) {
	value := operand.Get() - 1
	operand.Set(value)

	c.resetFlags(zero | halfCarry)
	c.setFlags(substract)

	if value == 0 {
		c.setFlags(zero)
	} else if value == 0xff {
		c.setFlags(halfCarry)
	}
}

func (c *CPU) ld(operand Operand, value byte) {
	operand.Set(value)
}

func (c *CPU) dec16(high, low *byte) {
	value := uint16(*high)<<8 | uint16(*low)

	value--

	*high = byte(value >> 8)
	*low = byte(value & 0xff)
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

func (c *CPU) ldi() {
	address := c.getAddress()
	c.a = c.mmu.ReadByte(address)

	address++

	c.h = byte(address >> 8)
	c.l = byte(address & 0xff)
}

// Takes the ones' complement of the contents of register A.
func (c *CPU) cpl() {
	c.a = ^c.a

	c.setFlags(substract | halfCarry)
}

func (c *CPU) std() {
	address := c.getAddress()
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
	c.pc = c.fetchWord()
}

func (c *CPU) ret() {
	c.pc = c.popStack()
}

func (c *CPU) call() {
	address := c.fetchWord()
	c.pushStack(c.pc)
	c.pc = address
}

func (c *CPU) sta8() {
	address := 0xff00 + uint16(c.fetch())
	c.mmu.WriteByte(address, c.a)
}

func (c *CPU) stc() {
	address := 0xff00 + uint16(c.c)
	c.mmu.WriteByte(address, c.a)
}

func (c *CPU) sta16() {
	address := c.fetchWord()
	c.mmu.WriteByte(address, c.a)
}

func (c *CPU) lda8() {
	address := 0xff00 + uint16(c.fetch())
	c.a = c.mmu.ReadByte(address)
}

func (c *CPU) di() {
	c.ime = false
}

func (c *CPU) ei() {
	c.ime = true
}

func (c *CPU) bit(bit byte, operand Operand) {
	c.resetFlags(substract)
	c.setFlags(halfCarry)

	value := operand.Get() >> bit & 1

	if value == 0 {
		c.setFlags(zero)
	} else {
		c.resetFlags(zero)
	}
}

func (c *CPU) res(bit byte, operand Operand) {
	value := operand.Get() &^ (1 << bit)
	operand.Set(value)
}

func (c *CPU) set(bit byte, operand Operand) {
	value := operand.Get() | 1<<bit
	operand.Set(value)
}
