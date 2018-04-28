package main

import (
	"flag"
	"log"

	"github.com/bovarysme/bmo/beemo"
	"github.com/bovarysme/bmo/debug"
)

var debugFlag bool
var bootromPath string
var romPath string
var screenScale int

func init() {
	flag.BoolVar(&debugFlag, "debug", false, "run the emulator in debug mode")
	flag.StringVar(&romPath, "rom", "", "path to the ROM file")
	flag.StringVar(&bootromPath, "bootrom", "roms/bootrom.gb", "path to the bootrom file")
	flag.IntVar(&screenScale, "scale", 2, "screen scale factor")

	flag.Parse()
}

func main() {
	bmo, err := beemo.NewBMO(bootromPath, romPath, screenScale)
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
