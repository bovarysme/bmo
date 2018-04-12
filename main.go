package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/bovarysme/bmo/beemo"
	"github.com/bovarysme/bmo/debug"
)

var debugFlag bool
var path string

func init() {
	flag.BoolVar(&debugFlag, "debug", false, "run the emulator in debug mode")
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

	if debugFlag {
		debugger := debug.NewDebugger(bmo)
		err = debugger.Run()
	} else {
		err = bmo.Run()
	}

	if err != nil {
		log.Fatal(err)
	}
}
