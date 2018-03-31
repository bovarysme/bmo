package main

import (
	"io/ioutil"
	"log"

	"github.com/bovarysme/bmo/cpu"
)

func main() {
	rom, err := ioutil.ReadFile("test.gb")
	if err != nil {
		log.Fatal(err)
	}

	c := cpu.NewCPU(rom)
	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}
}
