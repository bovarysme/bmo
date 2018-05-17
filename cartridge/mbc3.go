package cartridge

type MBC3 struct {
	rom     []byte
	romBank byte

	ram        [][]byte
	ramEnabled bool
	ramBank    byte

	bankingMode byte
}

func NewMBC3(rom []byte, ramType byte) *MBC3 {
	return &MBC3{
		rom:     rom,
		romBank: 1,

		ram: initRAM(ramType),
	}
}

func (m *MBC3) ReadByte(address uint16) byte {
	var value byte

	switch {
	case address >= 0 && address <= 0x3fff:
		value = m.rom[address]
	case address >= 0x4000 && address <= 0x7fff:
		addr := uint32(address) + 0x4000*(uint32(m.romBank)-1)
		value = m.rom[addr]
	case address >= 0xa000 && address <= 0xbfff:
		if m.ramEnabled {
			address -= 0xa000
			value = m.ram[m.ramBank][address]
		}
	}

	return value
}

func (m *MBC3) WriteByte(address uint16, value byte) {
	switch {
	case address >= 0 && address <= 0x1fff:
		m.ramEnabled = value&0xa == 0xa
	case address >= 0x2000 && address <= 0x3fff:
		if value == 0 {
			value = 1
		}

		m.romBank = value & 0x7f
	case address >= 0x4000 && address <= 0x5fff:
		m.ramBank = value & 3
	case address >= 0xa000 && address <= 0xbfff:
		if m.ramEnabled {
			address -= 0xa000
			m.ram[m.ramBank][address] = value
		}
	}
}

func (m *MBC3) Save() error {
	// TODO

	return nil
}
