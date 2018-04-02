package ppu

import (
	"github.com/bovarysme/bmo/mmu"
)

// PPU registers' addresses
const (
	LCDC uint16 = 0xff40 + iota // LCD Control Register (R/W)
	STAT                        // LCDC Status Flag (R/W)
	SCY                         // Scroll Y (R/W)
	SCX                         // Scroll X (R/W)
	LY                          // LCDC Y-Coordinate (R)
	LYC                         // LY Compare (R/W)
	DMA                         // DMA Transfer and Starting Address (W)
	BGP                         // BG Palette Data (W)
	OBP0                        // OBJ Palette Data 0 (W)
	OBP1                        // OBJ Palette Data 1 (W)
	WY                          // Window Y-Coordinate (R/W)
	WX                          // Window X-Coordinate (R/W)
)

// LCD Control register's masks
const (
	BGEnable byte = 1 << iota
	OBJEnable
	OBJSize
	BGTileMapAddress
	BGCharacterDataAddress
	WindowEnable
	WindowTileMapAddress
	LCDEnable
)

type PPU struct {
	mmu *mmu.MMU
}

func NewPPU(mmu *mmu.MMU) *PPU {
	return &PPU{
		mmu: mmu,
	}
}

func (p *PPU) Step() {
	if !p.getFlag(LCDC, LCDEnable) {
		return
	}

	value := p.mmu.ReadByte(LY)
	value++
	if value >= 154 {
		value = 0
	}

	p.mmu.WriteByte(LY, value)
}

func (p *PPU) getFlag(address uint16, mask byte) bool {
	value := p.mmu.ReadByte(LCDC)
	return value&mask == mask
}
