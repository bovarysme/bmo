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

// XXX
type MMU struct {
	ROM         []byte
	VideoRAM    [ExternalRAM - VideoRAM]byte
	ExternalRAM [RAM - ExternalRAM]byte
	RAM         [Forbidden - RAM]byte
	OAMRAM      [Unused - OAMRAM]byte
	IO          [HRAM - IO]byte
	HRAM        [RAMSize - HRAM]byte
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
	if address >= IO && address < HRAM {
		m.IO[address-IO] = value
	}
}
