package ppu

import (
	"github.com/bovarysme/bmo/interrupt"
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
	ColorDepth   = 4 // XXX: has to be 4 even with a RGB888 pixel format?

	Pitch      = ScreenWidth * ColorDepth
	bufferSize = ScreenWidth * ScreenHeight * ColorDepth
)

var Colors = [4][3]byte{
	{255, 255, 255}, // White
	{169, 169, 169}, // Light gray
	{54, 54, 54},    // Dark gray
	{0, 0, 0},       // Black
}

type PPU struct {
	Pixels []byte
	VBlank bool

	mode   byte
	cycles int

	ic  *interrupt.IC
	mmu *mmu.MMU
}

func NewPPU(mmu *mmu.MMU, ic *interrupt.IC) *PPU {
	return &PPU{
		Pixels: make([]byte, bufferSize),

		mode: OAMSearch,

		ic:  ic,
		mmu: mmu,
	}
}

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
			p.ic.Request(interrupt.VBlank)
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
func (p *PPU) transferLine(line byte) {
	if p.getFlag(LCDC, BGEnable) {
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

		mapLine := line + p.mmu.ReadByte(SCY)
		mapCol := p.mmu.ReadByte(SCX)

		// Tiles are 8 lines tall and maps 32 tiles wide (with one tile being
		// one byte)
		mapOffset := mapAddress + uint16(mapLine)/8*32

		paletteData := p.mmu.ReadByte(BGP)
		var palette [4]byte
		for i := uint(0); i < 4; i++ {
			palette[i] = paletteData >> (i * 2) & 0x3
		}

		// The screen is 20 tiles wide
		for i := 0; i < 20; i++ {
			var temp byte = (byte(i)*8 + mapCol) / 8
			address := mapOffset + uint16(temp)

			var tileNumber uint16
			if dataAddress == 0x8800 {
				tileNumber = uint16(int8(p.mmu.ReadByte(address))) + 128
			} else {
				tileNumber = uint16(p.mmu.ReadByte(address))
			}

			// Tile data = 16 bytes
			dataOffset := dataAddress + tileNumber*16 + uint16(mapLine)%8*2
			tileData := p.mmu.ReadWord(dataOffset)

			high := byte(tileData >> 8)
			low := byte(tileData & 0xff)

			for j := 0; j < 8; j++ {
				colorNumber := high>>(7-byte(j))&1<<1 | low>>(7-byte(j))&1
				color := Colors[palette[colorNumber]]

				x := i*8 + j

				index := Pitch*int(line) + ColorDepth*x
				p.Pixels[index] = color[0]
				p.Pixels[index+1] = color[1]
				p.Pixels[index+2] = color[2]
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
