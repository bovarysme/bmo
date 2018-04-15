package mmu

import (
	"errors"
	"io/ioutil"

	"github.com/bovarysme/bmo/cartridge"
)

const (
	romStart = 0
	romEnd   = 0x7fff
	romSize  = romEnd - romStart + 1

	vramStart = 0x8000
	vramEnd   = 0x9fff
	vramSize  = vramEnd - vramStart + 1

	externalRAMStart = 0xa000
	externalRAMEnd   = 0xbfff
	externalRAMSize  = externalRAMEnd - externalRAMStart + 1

	wramStart = 0xc000
	wramEnd   = 0xdfff
	wramSize  = wramEnd - wramStart + 1

	oamRAMStart = 0xfe00
	oamRAMEnd   = 0xfe9f
	oamRAMSize  = oamRAMEnd - oamRAMStart + 1

	ioStart = 0xff00
	ioEnd   = 0xff7f
	ioSize  = ioEnd - ioStart + 1

	hramStart = 0xff80
	hramEnd   = 0xffff
	hramSize  = hramEnd - hramStart + 1
)

const dmaRegisterAddress uint16 = 0xff46

type MMU struct {
	bootrom   []byte
	cartridge cartridge.Cartridge
	vram      [vramSize]byte
	wram      [wramSize]byte
	oamRAM    [oamRAMSize]byte
	io        [ioSize]byte
	hram      [hramSize]byte
}

func NewMMU(cartridge cartridge.Cartridge) (*MMU, error) {
	bootrom, err := ioutil.ReadFile("roms/bootrom.gb")
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

func (m *MMU) ReadByte(address uint16) byte {
	var value byte

	switch {
	case address >= romStart && address <= romEnd:
		if address < 0x100 && m.io[0x50] == 0 {
			value = m.bootrom[address]
		} else {
			value = m.cartridge.ReadByte(address)
		}
	case address >= vramStart && address <= vramEnd:
		address -= vramStart
		value = m.vram[address]

	case address >= externalRAMStart && address <= externalRAMEnd:
		value = m.cartridge.ReadByte(address)

	case address >= wramStart && address <= wramEnd:
		address -= wramStart
		value = m.wram[address]

	case address >= oamRAMStart && address <= oamRAMEnd:
		address -= oamRAMStart
		value = m.oamRAM[address]

	case address >= ioStart && address <= ioEnd:
		address -= ioStart
		value = m.io[address]

	case address >= hramStart && address <= hramEnd:
		address -= hramStart
		value = m.hram[address]
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

	case address >= vramStart && address <= vramEnd:
		address -= vramStart
		m.vram[address] = value

	case address >= externalRAMStart && address <= externalRAMEnd:
		m.cartridge.WriteByte(address, value)

	case address >= wramStart && address <= wramEnd:
		address -= wramStart
		m.wram[address] = value

	case address >= oamRAMStart && address <= oamRAMEnd:
		address -= oamRAMStart
		m.oamRAM[address] = value

	case address >= ioStart && address <= ioEnd:
		// XXX
		if address == 0xff00 {
			value |= 0xf
		}

		address -= ioStart
		m.io[address] = value

	case address >= hramStart && address <= hramEnd:
		address -= hramStart
		m.hram[address] = value
	}

	if address == dmaRegisterAddress {
		m.handleDMA(value)
	}
}

func (m *MMU) WriteWord(address, value uint16) {
	m.WriteByte(address, byte(value&0xff))
	m.WriteByte(address+1, byte(value>>8))
}

func (m *MMU) handleDMA(value byte) {
	source := uint16(value) << 8
	dest := uint16(oamRAMStart)

	for i := 0; i < 0xa0; i++ {
		b := m.ReadByte(source)
		m.WriteByte(dest, b)

		source++
		dest++
	}
}
