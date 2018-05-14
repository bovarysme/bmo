package mmu

import (
	"errors"
	"io/ioutil"

	"github.com/bovarysme/bmo/interrupt"
)

const (
	romStart = 0
	romEnd   = 0x7fff
	romSize  = romEnd - romStart + 1

	VRAMStart = 0x8000
	VRAMEnd   = 0x9fff
	VRAMSize  = VRAMEnd - VRAMStart + 1

	externalRAMStart = 0xa000
	externalRAMEnd   = 0xbfff
	externalRAMSize  = externalRAMEnd - externalRAMStart + 1

	wramStart = 0xc000
	wramEnd   = 0xdfff
	wramSize  = wramEnd - wramStart + 1

	OAMRAMStart = 0xfe00
	OAMRAMEnd   = 0xfe9f
	OAMRAMSize  = OAMRAMEnd - OAMRAMStart + 1

	ioStart = 0xff00
	ioEnd   = 0xff7f
	ioSize  = ioEnd - ioStart + 1

	hramStart = 0xff80
	hramEnd   = 0xfffe
	hramSize  = hramEnd - hramStart + 1
)

const dmaRegisterAddress uint16 = 0xff46

type Memory interface {
	ReadByte(address uint16) byte
	WriteByte(address uint16, value byte)
}

type MMU struct {
	bootrom   []byte
	cartridge Memory
	ic        Memory
	joypad    Memory
	ppu       Memory
	timer     Memory

	wram [wramSize]byte
	io   [ioSize]byte
	hram [hramSize]byte
}

func NewMMU(bootromPath string, cartridge Memory) (*MMU, error) {
	bootrom, err := ioutil.ReadFile(bootromPath)
	if err != nil {
		return nil, err
	}

	if len(bootrom) != 256 {
		return nil, errors.New("Invalid bootrom size")
	}

	return &MMU{
		bootrom:   bootrom,
		cartridge: cartridge,
	}, nil
}

// XXX
func (m *MMU) LinkIC(ic Memory) {
	m.ic = ic
}

func (m *MMU) LinkJoypad(joypad Memory) {
	m.joypad = joypad
}

func (m *MMU) LinkPPU(ppu Memory) {
	m.ppu = ppu
}

func (m *MMU) LinkTimer(timer Memory) {
	m.timer = timer
}

func (m *MMU) ReadByte(address uint16) byte {
	var value byte

	switch {
	case address >= romStart && address <= romEnd:
		if address < 0x100 && m.io[0x50] == 0 {
			value = m.bootrom[address]
		} else {
			value = m.cartridge.ReadByte(address)
		}
	case address >= VRAMStart && address <= VRAMEnd:
		value = m.ppu.ReadByte(address)

	case address >= externalRAMStart && address <= externalRAMEnd:
		value = m.cartridge.ReadByte(address)

	case address >= wramStart && address <= wramEnd:
		address -= wramStart
		value = m.wram[address]

	case address >= OAMRAMStart && address <= OAMRAMEnd:
		value = m.ppu.ReadByte(address)

	case address >= ioStart && address <= ioEnd:
		// XXX
		switch address {
		case 0xff00:
			value = m.joypad.ReadByte(address)
		case 0xff07:
			value = m.timer.ReadByte(address)
		case interrupt.IR:
			value = m.ic.ReadByte(address)
		case 0xff44:
			value = m.ppu.ReadByte(address)
		default:
			address -= ioStart
			value = m.io[address]
		}

	case address >= hramStart && address <= hramEnd:
		address -= hramStart
		value = m.hram[address]

	case address == interrupt.IE:
		value = m.ic.ReadByte(address)
	}

	return value
}

func (m *MMU) ReadWord(address uint16) uint16 {
	return uint16(m.ReadByte(address+1))<<8 | uint16(m.ReadByte(address))
}

func (m *MMU) WriteByte(address uint16, value byte) {
	switch {
	case address >= romStart && address <= romEnd:
		m.cartridge.WriteByte(address, value)

	case address >= VRAMStart && address <= VRAMEnd:
		m.ppu.WriteByte(address, value)

	case address >= externalRAMStart && address <= externalRAMEnd:
		m.cartridge.WriteByte(address, value)

	case address >= wramStart && address <= wramEnd:
		address -= wramStart
		m.wram[address] = value

	case address >= OAMRAMStart && address <= OAMRAMEnd:
		m.ppu.WriteByte(address, value)

	case address >= ioStart && address <= ioEnd:
		// XXX
		switch address {
		case 0xff00:
			m.joypad.WriteByte(address, value)
		case 0xff07:
			m.timer.WriteByte(address, value)
		case interrupt.IR:
			m.ic.WriteByte(address, value)
		case 0xff44:
			m.ppu.WriteByte(address, value)
		default:
			if address == dmaRegisterAddress {
				m.handleDMA(value)
			}

			address -= ioStart
			m.io[address] = value
		}

	case address >= hramStart && address <= hramEnd:
		address -= hramStart
		m.hram[address] = value

	case address == interrupt.IE:
		m.ic.WriteByte(address, value)
	}
}

func (m *MMU) WriteWord(address, value uint16) {
	m.WriteByte(address, byte(value&0xff))
	m.WriteByte(address+1, byte(value>>8))
}

func (m *MMU) handleDMA(value byte) {
	source := uint16(value) << 8
	dest := uint16(OAMRAMStart)

	for i := 0; i < 0xa0; i++ {
		b := m.ReadByte(source)
		m.WriteByte(dest, b)

		source++
		dest++
	}
}
