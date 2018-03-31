package main

import (
	"io/ioutil"
	"log"

	"github.com/bovarysme/bmo/cpu"
	"github.com/bovarysme/bmo/mmu"
)

func main() {
	rom, err := ioutil.ReadFile("test.gb")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("ROM size: %d bytes\n", len(rom))

	m := mmu.NewMMU(rom)

	c := cpu.NewCPU(m)
	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}
}
