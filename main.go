package main

import (
	"io/ioutil"
	"log"

	"github.com/bovarysme/bmo/cpu"
	"github.com/bovarysme/bmo/mmu"
	"github.com/bovarysme/bmo/ppu"
)

func main() {
	rom, err := ioutil.ReadFile("test.gb")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("ROM size: %d bytes\n", len(rom))

	m := mmu.NewMMU(rom)

	c := cpu.NewCPU(m)
	p := ppu.NewPPU(m)

	for {
		err = c.Step()
		if err != nil {
			log.Fatal(err)
		}

		p.Step()
	}
}
