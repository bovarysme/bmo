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

const (
	tileWidth     = 8
	tileHeight    = 8
	tileMapWidth  = 32
	tileMapHeight = 32
)

var Colors = [4][3]byte{
	{255, 255, 255}, // White
	{169, 169, 169}, // Light gray
	{54, 54, 54},    // Dark gray
	{0, 0, 0},       // Black
}

type Sprite struct {
	x           byte
	y           byte
	tileNumber  byte
	palette     uint16
	hFlip       bool
	vFlip       bool
	hasPriority bool
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

	cycles  int
	mode    byte
	sprites []Sprite
}

func NewPPU(mmu *mmu.MMU, ic *interrupt.IC) *PPU {
	return &PPU{
		Pixels: make([]byte, bufferSize),

		mode: OAMSearch,

		ic:  ic,
		mmu: mmu,

		sprites: make([]Sprite, 0, 10),
	}
}

func (p *PPU) Step(cycles int) {
	if !p.hasFlags(LCDC, LCDEnable) {
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
		case OAMSearch:
			p.oamSearch()
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

func (p *PPU) oamSearch() {
	p.sprites = p.sprites[:0]

	var spriteHeight byte = 8
	if p.hasFlags(LCDC, OBJSize) {
		spriteHeight = 16
	}

	for address := uint16(mmu.OAMRAMStart); address <= mmu.OAMRAMEnd; address += 4 {
		y := p.ReadByte(address)
		x := p.ReadByte(address + 1)

		if x == 0 || p.ly+16 < y || p.ly+16 >= y+spriteHeight {
			continue
		}

		tileNumber := p.ReadByte(address + 2)
		flags := p.ReadByte(address + 3)

		var palette uint16 = OBP0
		if flags>>4&1 == 1 {
			palette = OBP1
		}

		sprite := Sprite{
			x:           x,
			y:           y,
			tileNumber:  tileNumber,
			palette:     palette,
			hFlip:       flags>>5&1 == 1,
			vFlip:       flags>>6&1 == 1,
			hasPriority: flags>>7&1 == 0,
		}

		p.sprites = append(p.sprites, sprite)
		if len(p.sprites) >= 10 {
			break
		}
	}
}

// TODO: clean up
func (p *PPU) transferLine() {
	if p.hasFlags(LCDC, BGEnable) {
		mapLine := p.ly + p.mmu.ReadByte(SCY)
		mapCol := int(p.mmu.ReadByte(SCX))

		p.transferBGOrWindow(BGTileMapAddress, mapLine, mapCol, 0)
	}

	if p.hasFlags(LCDC, WindowEnable) {
		wy := p.mmu.ReadByte(WY)
		if p.ly >= wy {
			mapLine := p.ly - wy
			start := int(p.mmu.ReadByte(WX) - 7)
			p.transferBGOrWindow(WindowTileMapAddress, mapLine, -start, start)
		}
	}

	if p.hasFlags(LCDC, OBJEnable) {
		p.transferSprite()
	}
}

// TODO: clean up, rename
func (p *PPU) transferBGOrWindow(mapMask, mapLine byte, mapColumn, start int) {
	var dataAddress uint16 = 0x8800
	if p.hasFlags(LCDC, BGWindowTileDataAddress) {
		dataAddress = 0x8000
	}

	var mapAddress uint16 = 0x9800
	if p.hasFlags(LCDC, mapMask) {
		mapAddress = 0x9c00
	}

	// Tiles are 8 lines tall and maps 32 tiles wide (with one tile being
	// one byte)
	mapOffset := mapAddress + uint16(mapLine)/tileHeight*tileMapWidth

	palette := p.decodePalette(BGP)

	for i := start; i < ScreenWidth; i += tileWidth {
		// XXX
		var temp byte = byte(i) + byte(mapColumn)
		address := mapOffset + uint16(temp)/tileWidth

		var tileNumber uint16 = uint16(p.mmu.ReadByte(address))
		if dataAddress == 0x8800 {
			tileNumber = uint16(int8(tileNumber)) + 128
		}

		// Tile data = 16 bytes
		dataOffset := dataAddress + tileNumber*16 + uint16(mapLine)%tileHeight*2

		low := p.ReadByte(dataOffset)
		high := p.ReadByte(dataOffset + 1)

		for j := 0; j < tileWidth; j++ {
			x := i + j
			if x >= ScreenWidth {
				break
			}

			// TODO: simplify
			colorNumber := high>>(7-byte(j))&1<<1 | low>>(7-byte(j))&1
			color := Colors[palette[colorNumber]]

			index := Pitch*int(p.ly) + ColorDepth*x
			p.Pixels[index] = color[0]
			p.Pixels[index+1] = color[1]
			p.Pixels[index+2] = color[2]
		}
	}
}

// TODO: clean up, rename
func (p *PPU) transferSprite() {
	var spriteHeight byte = 8
	if p.hasFlags(LCDC, OBJSize) {
		spriteHeight = 16
	}

	for _, sprite := range p.sprites {
		palette := p.decodePalette(sprite.palette)

		// XXX
		dataOffset := 0x8000 + uint16(sprite.tileNumber)*16
		if sprite.vFlip {
			dataOffset += uint16((spriteHeight-1)-(p.ly-sprite.y)%spriteHeight) * 2
		} else {
			dataOffset += uint16((p.ly-sprite.y)%spriteHeight) * 2
		}

		low := p.ReadByte(dataOffset)
		high := p.ReadByte(dataOffset + 1)

		for j := 0; j < 8; j++ {
			// XXX
			x := int(sprite.x - 8)
			if sprite.hFlip {
				x += 7 - j
			} else {
				x += j
			}

			if x >= ScreenWidth {
				break
			}

			colorNumber := high>>(7-byte(j))&1<<1 | low>>(7-byte(j))&1
			if colorNumber == 0 {
				continue
			}

			color := Colors[palette[colorNumber]]

			index := Pitch*int(p.ly) + ColorDepth*x
			p.Pixels[index] = color[0]
			p.Pixels[index+1] = color[1]
			p.Pixels[index+2] = color[2]
		}
	}
}

// TODO: move to the MMU?
func (p *PPU) hasFlags(address uint16, mask byte) bool {
	value := p.mmu.ReadByte(address)
	return value&mask == mask
}

func (p *PPU) decodePalette(address uint16) [4]byte {
	data := p.mmu.ReadByte(address)

	var palette [4]byte
	for i := uint(0); i < 4; i++ {
		palette[i] = data >> (i * 2) & 0x3
	}

	return palette
}
