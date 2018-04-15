package mmu

import (
	"errors"
	"io/ioutil"

	"github.com/bovarysme/bmo/cartridge"
)

const (
	ROM         = 0x0
	VideoRAM    = 0x8000
	ExternalRAM = 0xa000
	RAM         = 0xc000
	Forbidden   = 0xe000
	OAMRAM      = 0xfe00
	Unused      = 0xfea0
	IO          = 0xff00
	HRAM        = 0xff80
	RAMSize     = 0xffff
)

const DMARegisterAddress uint16 = 0xff46

// XXX
type MMU struct {
	bootrom   []byte
	cartridge cartridge.Cartridge
	VideoRAM  [ExternalRAM - VideoRAM]byte
	RAM       [Forbidden - RAM]byte
	OAMRAM    [Unused - OAMRAM]byte
	IO        [HRAM - IO]byte
	HRAM      [0x10000 - HRAM]byte
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

// XXX
func (m *MMU) ReadByte(address uint16) byte {
	if address >= ROM && address < VideoRAM {
		bootrom := m.ReadByte(0xff50)
		if address < 0x100 && bootrom == 0 {
			return m.bootrom[address]
		} else {
			return m.cartridge.ReadByte(address)
		}
	} else if address >= VideoRAM && address < ExternalRAM {
		return m.VideoRAM[address-VideoRAM]
	} else if address >= ExternalRAM && address < RAM {
		return m.cartridge.ReadByte(address)
	} else if address >= RAM && address < Forbidden {
		return m.RAM[address-RAM]
	} else if address >= OAMRAM && address < Unused {
		return m.OAMRAM[address-OAMRAM]
	} else if address >= IO && address < HRAM {
		return m.IO[address-IO]
	} else if address >= HRAM && address <= RAMSize {
		return m.HRAM[address-HRAM]
	}

	return 0
}

func (m *MMU) ReadWord(address uint16) uint16 {
	return uint16(m.ReadByte(address+1))<<8 | uint16(m.ReadByte(address))
}

// XXX
func (m *MMU) WriteByte(address uint16, value byte) {
	if address >= ROM && address < VideoRAM {
		m.cartridge.WriteByte(address, value)
	} else if address >= VideoRAM && address < ExternalRAM {
		m.VideoRAM[address-VideoRAM] = value
	} else if address >= ExternalRAM && address < RAM {
		m.cartridge.WriteByte(address, value)
	} else if address >= RAM && address < Forbidden {
		m.RAM[address-RAM] = value
	} else if address >= OAMRAM && address < Unused {
		m.OAMRAM[address-OAMRAM] = value
	} else if address >= IO && address < HRAM {
		m.IO[address-IO] = value
	} else if address >= HRAM && address <= RAMSize {
		m.HRAM[address-HRAM] = value
	}

	if address == DMARegisterAddress {
		m.handleDMA(value)
	}
}

func (m *MMU) WriteWord(address, value uint16) {
	m.WriteByte(address, byte(value&0xff))
	m.WriteByte(address+1, byte(value>>8))
}

func (m *MMU) handleDMA(value byte) {
	source := uint16(value) << 8
	dest := uint16(OAMRAM)

	for i := 0; i < 0xa0; i++ {
		b := m.ReadByte(source)
		m.WriteByte(dest, b)

		source++
		dest++
	}
}
