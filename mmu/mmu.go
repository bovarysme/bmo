package mmu

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
	ROM         []byte
	VideoRAM    [ExternalRAM - VideoRAM]byte
	ExternalRAM [RAM - ExternalRAM]byte
	RAM         [Forbidden - RAM]byte
	OAMRAM      [Unused - OAMRAM]byte
	IO          [HRAM - IO]byte
	HRAM        [0x10000 - HRAM]byte
}

func NewMMU(rom []byte) *MMU {
	return &MMU{
		ROM: rom,
	}
}

// XXX
func (m *MMU) ReadByte(address uint16) byte {
	if address >= ROM && address < VideoRAM {
		return m.ROM[address]
	} else if address >= VideoRAM && address < ExternalRAM {
		return m.VideoRAM[address-VideoRAM]
	} else if address >= ExternalRAM && address < RAM {
		return m.ExternalRAM[address-ExternalRAM]
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
	if address >= VideoRAM && address < ExternalRAM {
		m.VideoRAM[address-VideoRAM] = value
	} else if address >= ExternalRAM && address < RAM {
		m.ExternalRAM[address-ExternalRAM] = value
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
