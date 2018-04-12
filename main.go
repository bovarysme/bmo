package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/bovarysme/bmo/beemo"
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

	bmo, err := beemo.NewBMO(rom)
	if err != nil {
		log.Fatal(err)
	}

	err = bmo.Run()
	if err != nil {
		log.Fatal(err)
	}
}
