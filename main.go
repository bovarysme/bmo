package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/bovarysme/bmo/cpu"
	"github.com/bovarysme/bmo/mmu"
	"github.com/bovarysme/bmo/ppu"
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

	for {
		cycles, err := c.Step()
		if err != nil {
			log.Fatal(err)
		}

		p.Step(cycles)
	}
}
