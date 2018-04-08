package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/bovarysme/bmo/cpu"
	"github.com/bovarysme/bmo/mmu"
	"github.com/bovarysme/bmo/ppu"

	"github.com/veandco/go-sdl2/sdl"
)

var path string

func init() {
	flag.StringVar(&path, "path", "", "path to the ROM file")

	flag.Parse()
}

func main() {
	rom, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("ROM size: %d bytes\n", len(rom))

	m := mmu.NewMMU(rom)

	c := cpu.NewCPU(m)
	p := ppu.NewPPU(m)

	err = sdl.Init(sdl.INIT_VIDEO)
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("BMO", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		ppu.ScreenWidth, ppu.ScreenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatal(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGB888, sdl.TEXTUREACCESS_STREAMING,
		ppu.ScreenWidth, ppu.ScreenHeight)
	if err != nil {
		log.Fatal(err)
	}
	defer texture.Destroy()

	renderer.Clear()

	for {
		cycles, err := c.Step()
		if err != nil {
			log.Fatal(err)
		}

		p.Step(cycles)

		if p.VBlank {
			p.VBlank = false

			for y := 0; y < ppu.ScreenHeight; y++ {
				for x := 0; x < ppu.ScreenWidth; x++ {
					colorIndex := p.Screen[y][x]
					color := ppu.Colors[colorIndex]

					err = renderer.SetDrawColor(color[0], color[1], color[2], 255)
					if err != nil {
						log.Fatal(err)
					}

					err = renderer.DrawPoint(x, y)
					if err != nil {
						log.Fatal(err)
					}
				}
			}

			renderer.Present()
		}
	}
}
