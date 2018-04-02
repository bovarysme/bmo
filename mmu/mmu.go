package mmu

type MMU struct {
	rom []byte
}

func NewMMU(rom []byte) *MMU {
	return &MMU{
		rom: rom,
	}
}

func (m *MMU) ReadByte(address uint16) byte {
	return m.rom[address]
}

func (m *MMU) ReadWord(address uint16) uint16 {
	return uint16(m.rom[address+1])<<8 | uint16(m.rom[address])
}

func (m *MMU) WriteByte(address uint16, value byte) {

}
