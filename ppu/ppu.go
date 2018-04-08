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
	BGTileDataAddress
	WindowEnable
	WindowTileMapAddress
	LCDEnable
)

// STAT register's masks
const (
	Mode byte = 0x3
)

const (
	HBlank byte = iota
	VBlank
	OAMSearch
	PixelTransfer
)

const (
	ScreenWidth  = 160
	ScreenHeight = 144
)

var Colors = map[byte][]byte{
	0: []byte{255, 255, 255}, // White
	3: []byte{128, 128, 128}, // Light gray
	2: []byte{64, 64, 64},    // Dark gray
	1: []byte{0, 0, 0},       // Black
}

type PPU struct {
	Screen [ScreenHeight][ScreenWidth]byte
	VBlank bool

	mode   byte
	cycles int

	mmu *mmu.MMU
}

func NewPPU(mmu *mmu.MMU) *PPU {
	return &PPU{
		mode: OAMSearch,

		mmu: mmu,
	}
}

/* Timings:
OAM Search: 20 clocks
Pixel Transfer: 43+ clocks (if displaying sprites or a window)
H-Blank: 51 clocks

114 x 154 = 17,556 clocks per screen
(w/ 10 lines of V-Blank)

Pixel Transfer: no CPU access to VRAM
OAM Search & Pixel transfer: no CPU access to OAM RAM
*/
func (p *PPU) Step(cycles int) {
	if !p.getFlag(LCDC, LCDEnable) {
		return
	}

	line := p.mmu.ReadByte(LY)
	p.updateMode(line)
	p.updateLine(line)

	p.cycles += cycles
}

func (p *PPU) updateMode(line byte) {
	var mode byte

	if line >= ScreenHeight {
		mode = VBlank
	} else if p.cycles >= 0 && p.cycles < 20 {
		mode = OAMSearch
	} else if p.cycles >= 20 && p.cycles < 63 {
		mode = PixelTransfer
	} else if p.cycles >= 63 && p.cycles < 114 {
		mode = HBlank
	}

	if mode != p.mode {
		p.mode = mode

		stat := p.mmu.ReadByte(STAT)
		stat = stat&^Mode | mode
		p.mmu.WriteByte(STAT, stat)

		switch mode {
		case PixelTransfer:
			p.transferLine(line)
		case VBlank:
			p.VBlank = true
			p.requestInterrupt()
		}
	}
}

func (p *PPU) updateLine(line byte) {
	if p.cycles >= 114 {
		p.cycles = 0

		line++
		if line >= 154 {
			line = 0
		}

		p.mmu.WriteByte(LY, line)
	}
}

// TODO: clean up
func (p *PPU) requestInterrupt() {
	address := uint16(0xff0f)

	ir := p.mmu.ReadByte(address)
	ir |= 1
	p.mmu.WriteByte(address, ir)
}

func (p *PPU) transferLine(line byte) {
	if p.getFlag(LCDC, BGEnable) {
		mapLine := line + p.mmu.ReadByte(SCY)

		var dataAddress uint16
		if p.getFlag(LCDC, BGTileDataAddress) {
			dataAddress = 0x8000
		} else {
			dataAddress = 0x8800
		}

		var mapAddress uint16
		if p.getFlag(LCDC, BGTileMapAddress) {
			mapAddress = 0x9c00
		} else {
			mapAddress = 0x9800
		}

		// Tiles are 8 lines tall and maps 32 tiles wide (with one tile being
		// one byte)
		mapOffset := mapAddress + uint16(mapLine)/8*32

		// The screen is 20 tiles wide
		for i := 0; i < 20; i++ {
			address := mapOffset + uint16(i)
			tileNumber := p.mmu.ReadByte(address)

			// Tile data = 16 bytes
			dataOffset := dataAddress + uint16(tileNumber)*16 + uint16(mapLine)%8*2
			tileData := p.mmu.ReadWord(dataOffset)

			high := byte(tileData >> 8)
			low := byte(tileData & 0xff)

			for j := 0; j < 8; j++ {
				color := high>>(7-byte(j))&1 | low>>(7-byte(j))&1

				x := i*8 + j
				p.Screen[line][x] = color
			}
		}
	}

	if p.getFlag(LCDC, OBJEnable) {
		// TODO
	}

	if p.getFlag(LCDC, WindowEnable) {
		// TODO
	}
}

// TODO: move to the MMU?
func (p *PPU) getFlag(address uint16, mask byte) bool {
	value := p.mmu.ReadByte(address)
	return value&mask == mask
}
