package cartridge

const (
	romBanking byte = iota
	ramBanking
)

type MBC1 struct {
	rom     []byte
	romBank byte

	ram        [][]byte
	ramEnabled bool
	ramBank    byte

	bankingMode byte
}

func NewMBC1(rom []byte, ramType byte) *MBC1 {
	return &MBC1{
		rom:     rom,
		romBank: 1,

		ram: initRAM(ramType),
	}
}

func (m *MBC1) ReadByte(address uint16) byte {
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

func (m *MBC1) WriteByte(address uint16, value byte) {
	switch {
	case address >= 0 && address <= 0x1fff:
		m.ramEnabled = value&0xa == 0xa
	case address >= 0x2000 && address <= 0x3fff:
		m.updateROMBank(value, 0x1f)
	case address >= 0x4000 && address <= 0x5fff:
		if m.bankingMode == romBanking {
			m.updateROMBank(value, 0x60)
		} else {
			m.ramBank = value & 3
		}
	case address >= 0x6000 && address <= 0x7fff:
		m.bankingMode = value & 1
	case address >= 0xa000 && address <= 0xbfff:
		if m.ramEnabled {
			address -= 0xa000
			m.ram[m.ramBank][address] = value
		}
	}
}

func (m *MBC1) updateROMBank(value, mask byte) {
	m.romBank &^= mask
	m.romBank |= value & mask

	if m.romBank == 0 || m.romBank == 0x20 || m.romBank == 0x40 || m.romBank == 0x60 {
		m.romBank++
	}
}
