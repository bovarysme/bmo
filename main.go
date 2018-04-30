package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"github.com/bovarysme/bmo/beemo"
	"github.com/bovarysme/bmo/debug"
)

var debugFlag bool
var screenScale int
var cpuprofile string
var bootromPath string
var romPath string

func init() {
	flag.BoolVar(&debugFlag, "debug", false, "run the emulator in debug mode")
	flag.IntVar(&screenScale, "scale", 2, "screen scale factor")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write a CPU profile")
	flag.StringVar(&romPath, "rom", "", "path to the ROM file")
	flag.StringVar(&bootromPath, "bootrom", "roms/bootrom.gb", "path to the bootrom file")

	flag.Parse()
}

func main() {
	if cpuprofile != "" {
		file, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}

		err = pprof.StartCPUProfile(file)
		if err != nil {
			log.Fatal(err)
		}

		defer pprof.StopCPUProfile()
	}

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
