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
	BGWindowTileDataAddress
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

	ic  *interrupt.IC
	mmu *mmu.MMU

	vram   [mmu.VRAMSize]byte
	oamRAM [mmu.OAMRAMSize]byte

	// Current line
	ly byte

	mode   byte
	cycles int
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

	p.updateMode()
	p.updateLine()

	p.cycles += cycles
}

func (p *PPU) ReadByte(address uint16) byte {
	var value byte

	switch {
	case address >= mmu.VRAMStart && address <= mmu.VRAMEnd:
		address -= mmu.VRAMStart
		value = p.vram[address]

	case address >= mmu.OAMRAMStart && address <= mmu.OAMRAMEnd:
		address -= mmu.OAMRAMStart
		value = p.oamRAM[address]

	case address == LY:
		value = p.ly
	}

	return value
}

func (p *PPU) WriteByte(address uint16, value byte) {
	switch {
	case address >= mmu.VRAMStart && address <= mmu.VRAMEnd:
		address -= mmu.VRAMStart
		p.vram[address] = value

	case address >= mmu.OAMRAMStart && address <= mmu.OAMRAMEnd:
		address -= mmu.OAMRAMStart
		p.oamRAM[address] = value
	}
}

func (p *PPU) updateMode() {
	var mode byte

	if p.ly >= ScreenHeight {
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
			p.transferLine()
		case VBlank:
			p.VBlank = true
			p.ic.Request(interrupt.VBlank)
		}
	}
}

func (p *PPU) updateLine() {
	if p.cycles >= 114 {
		p.cycles = 0

		p.ly++
		if p.ly >= 154 {
			p.ly = 0
		}
	}
}

// TODO: clean up, rename
func (p *PPU) transferBGOrWindow(mapMask, mapLine byte, mapColumn, start int) {
	const (
		tileWidth    = 8
		tileHeight   = 8
		tileMapWidth = 32
	)

	var dataAddress, mapAddress uint16
	if p.getFlag(LCDC, BGWindowTileDataAddress) {
		dataAddress = 0x8000
	} else {
		dataAddress = 0x8800
	}

	if p.getFlag(LCDC, mapMask) {
		mapAddress = 0x9c00
	} else {
		mapAddress = 0x9800
	}

	// Tiles are 8 lines tall and maps 32 tiles wide (with one tile being
	// one byte)
	mapOffset := mapAddress + uint16(mapLine)/tileHeight*tileMapWidth

	paletteData := p.mmu.ReadByte(BGP)
	var palette [4]byte
	for i := uint(0); i < 4; i++ {
		palette[i] = paletteData >> (i * 2) & 0x3
	}

	for i := start; i < ScreenWidth; i += tileWidth {
		// XXX
		var temp byte = byte(i) + byte(mapColumn)
		address := mapOffset + uint16(temp)/tileWidth

		var tileNumber uint16
		if dataAddress == 0x8800 {
			tileNumber = uint16(int8(p.mmu.ReadByte(address))) + 128
		} else {
			tileNumber = uint16(p.mmu.ReadByte(address))
		}

		// Tile data = 16 bytes
		dataOffset := dataAddress + tileNumber*16 + uint16(mapLine)%tileHeight*2
		tileData := p.mmu.ReadWord(dataOffset)

		high := byte(tileData >> 8)
		low := byte(tileData & 0xff)

		for j := 0; j < tileWidth; j++ {
			x := i + j
			if x >= ScreenWidth {
				break
			}

			colorNumber := high>>(7-byte(j))&1<<1 | low>>(7-byte(j))&1
			color := Colors[palette[colorNumber]]

			index := Pitch*int(p.ly) + ColorDepth*x
			p.Pixels[index] = color[0]
			p.Pixels[index+1] = color[1]
			p.Pixels[index+2] = color[2]
		}
	}
}

// TODO: clean up
func (p *PPU) transferLine() {
	if p.getFlag(LCDC, BGEnable) {
		mapLine := p.ly + p.mmu.ReadByte(SCY)
		mapCol := int(p.mmu.ReadByte(SCX))

		p.transferBGOrWindow(BGTileMapAddress, mapLine, mapCol, 0)
	}

	if p.getFlag(LCDC, WindowEnable) {
		wy := p.mmu.ReadByte(WY)
		if p.ly >= wy {
			mapLine := p.ly - wy
			start := int(p.mmu.ReadByte(WX) - 7)
			p.transferBGOrWindow(WindowTileMapAddress, mapLine, -start, start)
		}
	}

	if p.getFlag(LCDC, OBJEnable) {
		// TODO: render sprites
	}
}

// TODO: move to the MMU?
func (p *PPU) getFlag(address uint16, mask byte) bool {
	value := p.mmu.ReadByte(address)
	return value&mask == mask
}
