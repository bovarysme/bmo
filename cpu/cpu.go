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

const (
	interruptEnableAddress  uint16 = 0xffff
	interruptRequestAddress        = 0xff0f
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

var prefixCycles = [...]int{
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
	2, 2, 2, 2, 2, 2, 3, 2, 2, 2, 2, 2, 2, 2, 3, 2,
	2, 2, 2, 2, 2, 2, 3, 2, 2, 2, 2, 2, 2, 2, 3, 2,
	2, 2, 2, 2, 2, 2, 3, 2, 2, 2, 2, 2, 2, 2, 3, 2,
	2, 2, 2, 2, 2, 2, 3, 2, 2, 2, 2, 2, 2, 2, 3, 2,
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2,
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

type Register struct {
	register *byte
}

func (r *Register) Get() byte {
	return *r.register
}

func (r *Register) Set(value byte) {
	*r.register = value
}

type Memory struct {
	address uint16
	mmu     *mmu.MMU
}

func (m *Memory) Get() byte {
	return m.mmu.ReadByte(m.address)
}

func (m *Memory) Set(value byte) {
	m.mmu.WriteByte(m.address, value)
}

type ExtendedOperand interface {
	Get() uint16
	Set(value uint16)
}

type ExtendedRegister struct {
	high *byte
	low  *byte
}

func (e *ExtendedRegister) Get() uint16 {
	return uint16(*e.high)<<8 | uint16(*e.low)
}

func (e *ExtendedRegister) Set(value uint16) {
	*e.high = byte(value >> 8)
	*e.low = byte(value & 0xff)
}

type StackPointer struct {
	sp *uint16
}

func (s *StackPointer) Get() uint16 {
	return *s.sp
}

func (s *StackPointer) Set(value uint16) {
	*s.sp = value
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

func (c *CPU) Step() (int, error) {
	c.cycles = 0

	if c.ime {
		interrupted := c.handleInterrupts()
		if interrupted {
			return c.cycles, nil
		}
	}

	//fmt.Printf("%#v\n", c)
	opcode := c.fetch()
	//fmt.Printf("opcode: %#x\n\n", opcode)

	err := c.decode(opcode)

	return c.cycles, err
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
	case 0x01, 0x11, 0x21, 0x31:
		operand := c.getExtendedOperand(opcode)
		c.ld16(operand)
	case 0x02, 0x12:
		operand := c.getExtendedOperand(opcode)
		c.str16(operand)
	case 0x03, 0x13, 0x23, 0x33:
		operand := c.getExtendedOperand(opcode)
		c.inc16(operand)
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
	case 0x07:
		c.rlca()
	case 0x08:
		c.stspa16()
	case 0x09, 0x19, 0x29, 0x39:
		operand := c.getExtendedOperand(opcode)
		c.add16(operand)
	case 0x0a, 0x1a:
		operand := c.getExtendedOperand(opcode)
		c.ldr16(operand)
	case 0x0b, 0x1b, 0x2b, 0x3b:
		operand := c.getExtendedOperand(opcode)
		c.dec16(operand)
	case 0x17:
		c.rla()
	case 0x18:
		c.jr()
	case 0x1f:
		c.rra()
	case 0x20, 0x28, 0x30, 0x38:
		condition := c.decodeCondition(opcode)
		c.jrc(condition)
	case 0x22:
		c.sti()
	case 0x27:
		c.daa()
	case 0x2a:
		c.ldi()
	case 0x2f:
		c.cpl()
	case 0x32:
		c.std()
	case 0x37:
		c.scf()
	case 0x3f:
		c.ccf()
	case 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0xc6:
		value := c.getSourceValue(opcode)
		c.add(value)
	case 0x88, 0x89, 0x8a, 0x8b, 0x8c, 0x8d, 0x8e, 0x8f, 0xce:
		value := c.getSourceValue(opcode)
		c.adc(value)
	case 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0xd6:
		value := c.getSourceValue(opcode)
		c.sub(value)
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
	// TODO: merge with ret?
	case 0xc0, 0xc8, 0xd0, 0xd8:
		condition := c.decodeCondition(opcode)
		c.retc(condition)
	case 0xc1, 0xd1, 0xe1, 0xf1:
		high, low := c.decodeRegisterPair(opcode)
		c.pop(high, low)
	// TODO: merge with jp?
	case 0xc2, 0xca, 0xd2, 0xda:
		condition := c.decodeCondition(opcode)
		c.jpc(condition)
	case 0xc3:
		c.jp()
	// TODO: merge with call?
	case 0xc4, 0xcc, 0xd4, 0xdc:
		condition := c.decodeCondition(opcode)
		c.callc(condition)
	case 0xc5, 0xd5, 0xe5, 0xf5:
		operand := c.getExtendedOperand(opcode)
		c.push(operand)
	case 0xc7, 0xcf, 0xd7, 0xdf, 0xe7, 0xef, 0xf7, 0xff:
		address := c.decodeAddress(opcode)
		c.rst(address)
	case 0xc9:
		c.ret()
	case 0xcb:
		err := c.decodePrefix()
		if err != nil {
			return err
		}
	case 0xcd:
		c.call()
	case 0xd9:
		c.reti()
	case 0xe0:
		c.sta8()
	case 0xe2:
		c.stc()
	case 0xea:
		c.sta16()
	case 0xe9:
		c.jphl()
	case 0xf0:
		c.lda8()
	case 0xf3:
		c.di()
	case 0xf8:
		c.stspr8()
	case 0xf9:
		c.ldsp()
	case 0xfa:
		c.lda16()
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
	operand := c.getSourceOperand(opcode)

	switch {
	case opcode >= 0x00 && opcode <= 0x07:
		c.rlc(operand)
	case opcode >= 0x08 && opcode <= 0x0f:
		c.rrc(operand)
	case opcode >= 0x10 && opcode <= 0x17:
		c.rl(operand)
	case opcode >= 0x18 && opcode <= 0x1f:
		c.rr(operand)
	case opcode >= 0x20 && opcode <= 0x27:
		c.sla(operand)
	case opcode >= 0x30 && opcode <= 0x37:
		c.swap(operand)
	case opcode >= 0x38 && opcode <= 0x3f:
		c.srl(operand)
	case opcode >= 0x40 && opcode <= 0x7f:
		bit := c.decodeBit(opcode)
		c.bit(bit, operand)
	case opcode >= 0x80 && opcode <= 0xbf:
		bit := c.decodeBit(opcode)
		c.res(bit, operand)
	case opcode >= 0xc0 && opcode <= 0xff:
		bit := c.decodeBit(opcode)
		c.set(bit, operand)
	default:
		return &UnknownPrefixOpcodeError{opcode: opcode}
	}

	c.cycles += prefixCycles[opcode]

	return nil
}

func (c *CPU) handleInterrupts() bool {
	ie := c.mmu.ReadByte(interruptEnableAddress)
	ir := c.mmu.ReadByte(interruptRequestAddress)

	for i := 0; i <= 4; i++ {
		var mask byte = 1 << byte(i)

		enabled := ie&mask == mask
		requested := ir&mask == mask

		if enabled && requested {
			c.cycles += 5
			c.ime = false

			ir &^= mask
			c.mmu.WriteByte(interruptRequestAddress, ir)

			c.pushStack(c.pc)
			c.pc = 0x40 + uint16(i)*8

			return true
		}
	}

	return false
}

func (c *CPU) popStack() uint16 {
	value := c.mmu.ReadWord(c.sp)
	c.sp += 2

	return value
}

func (c *CPU) pushStack(value uint16) {
	c.mmu.WriteWord(c.sp-2, value)
	c.sp -= 2
}

func (c *CPU) decodeAddress(opcode byte) uint16 {
	return uint16(opcode >> 3 & 0x7 * 8)
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
	var high, low *byte

	switch opcode >> 4 & 0x3 {
	case 0:
		high, low = &c.b, &c.c
	case 1:
		high, low = &c.d, &c.e
	case 2:
		high, low = &c.h, &c.l
	case 3:
		high, low = &c.a, &c.f
	}

	return high, low
}

func (c *CPU) getHL() uint16 {
	return uint16(c.h)<<8 | uint16(c.l)
}

func (c *CPU) setHL(value uint16) {
	c.h = byte(value >> 8)
	c.l = byte(value & 0xff)
}

func (c *CPU) getOperand(register *byte) Operand {
	if register != nil {
		return &Register{register: register}
	}

	address := c.getHL()
	return &Memory{
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

func (c *CPU) getExtendedOperand(opcode byte) ExtendedOperand {
	if opcode>>4 == 0x3 {
		return &StackPointer{sp: &c.sp}
	}

	high, low := c.decodeRegisterPair(opcode)
	return &ExtendedRegister{
		high: high,
		low:  low,
	}
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
	address := c.getHL()
	return c.mmu.ReadByte(address)
}

func (c *CPU) getFlags(value byte) byte {
	var result byte
	if c.f&value == value {
		result = 1
	}

	return result
}

func (c *CPU) setFlags(value byte) {
	c.f |= value
}

func (c *CPU) resetFlags(value byte) {
	c.f &^= value
}

func (c *CPU) nop() {

}

func (c *CPU) ld16(operand ExtendedOperand) {
	value := c.fetchWord()
	operand.Set(value)
}

func (c *CPU) str16(operand ExtendedOperand) {
	address := operand.Get()
	c.mmu.WriteByte(address, c.a)
}

func (c *CPU) inc16(operand ExtendedOperand) {
	value := operand.Get()
	value++
	operand.Set(value)
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

func (c *CPU) rlca() {
	bit := c.a >> 7
	c.a = c.a<<1 | bit

	c.resetFlags(zero | substract | halfCarry | carry)
	if bit == 1 {
		c.setFlags(carry)
	}
}

func (c *CPU) stspa16() {
	address := c.fetchWord()
	c.mmu.WriteWord(address, c.sp)
}

func (c *CPU) add16(operand ExtendedOperand) {
	c.resetFlags(substract | halfCarry | carry)

	temp := uint32(c.getHL()) + uint32(operand.Get())
	// XXX
	if temp>>12&1 == 1 {
		c.setFlags(halfCarry)
	}
	if temp > 0xffff {
		c.setFlags(carry)
	}

	c.setHL(uint16(temp))
}

func (c *CPU) ldr16(operand ExtendedOperand) {
	address := operand.Get()
	c.a = c.mmu.ReadByte(address)
}

func (c *CPU) dec16(operand ExtendedOperand) {
	value := operand.Get()
	value--
	operand.Set(value)
}

func (c *CPU) rla() {
	bit := c.a >> 7
	c.a = c.a<<1 | c.getFlags(carry)

	c.resetFlags(zero | substract | halfCarry | carry)
	if bit == 1 {
		c.setFlags(carry)
	}
}

func (c *CPU) jr() {
	steps := int8(c.fetch())

	// XXX: is there a cleaner way to do this?
	c.pc = uint16(int16(c.pc) + int16(steps))
}

func (c *CPU) rra() {
	bit := c.a & 1
	c.a = c.getFlags(carry)<<7 | c.a>>1

	c.resetFlags(zero | substract | halfCarry | carry)
	if bit == 1 {
		c.setFlags(carry)
	}
}

func (c *CPU) jrc(condition bool) {
	steps := int8(c.fetch())

	if condition {
		c.cycles++

		// XXX: is there a cleaner way to do this?
		c.pc = uint16(int16(c.pc) + int16(steps))
	}
}

func (c *CPU) sti() {
	address := c.getHL()
	c.mmu.WriteByte(address, c.a)

	address++

	c.setHL(address)
}

func (c *CPU) daa() {

}

func (c *CPU) ldi() {
	address := c.getHL()
	c.a = c.mmu.ReadByte(address)

	address++

	c.setHL(address)
}

// Takes the ones' complement of the contents of register A.
func (c *CPU) cpl() {
	c.a = ^c.a

	c.setFlags(substract | halfCarry)
}

func (c *CPU) std() {
	address := c.getHL()
	c.mmu.WriteByte(address, c.a)

	address--

	c.setHL(address)
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

func (c *CPU) add(value byte) {
	c.resetFlags(zero | substract | halfCarry | carry)

	temp := uint16(c.a) + uint16(value)
	if temp>>4&1 == 1 {
		c.setFlags(halfCarry)
	}
	if temp > 0xff {
		c.setFlags(carry)
	}

	c.a = byte(temp)

	if c.a == 0 {
		c.setFlags(zero)
	}
}

func (c *CPU) adc(value byte) {
	cy := uint16(c.getFlags(carry))
	temp := uint16(c.a) + uint16(value) + cy

	c.resetFlags(zero | substract | halfCarry | carry)
	if temp>>4&1 == 1 {
		c.setFlags(halfCarry)
	}
	if temp > 0xff {
		c.setFlags(carry)
	}

	c.a = byte(temp)

	if c.a == 0 {
		c.setFlags(zero)
	}
}

func (c *CPU) sub(value byte) {
	c.resetFlags(zero | halfCarry | carry)
	c.setFlags(substract)

	if c.a&0x0f < value&0x0f {
		c.setFlags(halfCarry)
	}
	if c.a < value {
		c.setFlags(carry)
	}

	c.a -= value

	if c.a == 0 {
		c.setFlags(zero)
	}
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

func (c *CPU) retc(condition bool) {
	if condition {
		c.cycles += 3
		c.pc = c.popStack()
	}
}

func (c *CPU) pop(high, low *byte) {
	value := c.popStack()

	*high = byte(value >> 8)
	*low = byte(value & 0xff)

	// Ensures the lower bits of register F are zeros.
	if low == &c.f {
		c.f &= 0xf0
	}
}

func (c *CPU) jpc(condition bool) {
	address := c.fetchWord()

	if condition {
		c.cycles++
		c.pc = address
	}
}

func (c *CPU) jp() {
	c.pc = c.fetchWord()
}

func (c *CPU) callc(condition bool) {
	address := c.fetchWord()

	if condition {
		c.cycles += 3
		c.pushStack(c.pc)
		c.pc = address
	}
}

func (c *CPU) push(operand ExtendedOperand) {
	value := operand.Get()
	c.pushStack(value)
}

func (c *CPU) rst(address uint16) {
	c.pushStack(c.pc)
	c.pc = address
}

func (c *CPU) ret() {
	c.pc = c.popStack()
}

func (c *CPU) call() {
	address := c.fetchWord()
	c.pushStack(c.pc)
	c.pc = address
}

func (c *CPU) reti() {
	c.ime = true
	c.pc = c.popStack()
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

func (c *CPU) jphl() {
	c.pc = c.getHL()
}

func (c *CPU) lda8() {
	address := 0xff00 + uint16(c.fetch())
	c.a = c.mmu.ReadByte(address)
}

func (c *CPU) di() {
	c.ime = false
}

func (c *CPU) stspr8() {
	steps := int8(c.fetch())

	c.resetFlags(zero | substract | halfCarry | carry)
	temp := uint32(c.sp) + uint32(steps)
	// XXX
	if temp>>12&1 == 1 {
		c.setFlags(halfCarry)
	}
	if temp > 0xffff {
		c.setFlags(carry)
	}

	c.setHL(uint16(temp))
}

func (c *CPU) ldsp() {
	c.sp = c.getHL()
}

func (c *CPU) lda16() {
	address := c.fetchWord()
	c.a = c.mmu.ReadByte(address)
}

func (c *CPU) ei() {
	c.ime = true
}

func (c *CPU) rlc(operand Operand) {
	value := operand.Get()

	bit := value >> 7
	value = value<<1 | bit

	operand.Set(value)

	c.resetFlags(zero | substract | halfCarry | carry)
	if value == 0 {
		c.setFlags(zero)
	}
	if bit == 1 {
		c.setFlags(carry)
	}
}

func (c *CPU) rrc(operand Operand) {
	value := operand.Get()

	bit := value & 1
	value = bit<<7 | value>>1

	operand.Set(value)

	c.resetFlags(zero | substract | halfCarry | carry)
	if value == 0 {
		c.setFlags(zero)
	}
	if bit == 1 {
		c.setFlags(carry)
	}
}

func (c *CPU) rl(operand Operand) {
	value := operand.Get()

	bit := value >> 7
	value = value<<1 | c.getFlags(carry)

	operand.Set(value)

	c.resetFlags(zero | substract | halfCarry | carry)
	if value == 0 {
		c.setFlags(zero)
	}
	if bit == 1 {
		c.setFlags(carry)
	}
}

func (c *CPU) rr(operand Operand) {
	value := operand.Get()

	bit := value & 1
	value = c.getFlags(carry)<<7 | value>>1

	operand.Set(value)

	c.resetFlags(zero | substract | halfCarry | carry)
	if value == 0 {
		c.setFlags(zero)
	}
	if bit == 1 {
		c.setFlags(carry)
	}
}

func (c *CPU) sla(operand Operand) {
	value := operand.Get()

	bit := value >> 7
	value <<= 1

	operand.Set(value)

	c.resetFlags(zero | substract | halfCarry | carry)
	if value == 0 {
		c.setFlags(zero)
	}
	if bit == 1 {
		c.setFlags(carry)
	}
}

func (c *CPU) swap(operand Operand) {
	value := operand.Get()

	lower := value & 0x0f
	value = lower<<4 | value>>4

	operand.Set(value)

	c.resetFlags(zero | substract | halfCarry | carry)
	if value == 0 {
		c.setFlags(zero)
	}
}

func (c *CPU) srl(operand Operand) {
	value := operand.Get()

	bit := value & 1
	value >>= 1

	operand.Set(value)

	c.resetFlags(zero | substract | halfCarry | carry)
	if value == 0 {
		c.setFlags(zero)
	}
	if bit == 1 {
		c.setFlags(carry)
	}
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
