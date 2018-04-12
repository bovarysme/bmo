package debug

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bovarysme/bmo/beemo"
)

type Debugger struct {
	bmo *beemo.BMO

	reader  *bufio.Reader
	running bool

	breaking   bool
	breakpoint uint16
}

func NewDebugger(bmo *beemo.BMO) *Debugger {
	return &Debugger{
		bmo: bmo,

		reader:  bufio.NewReader(os.Stdin),
		running: true,
	}
}

func (d *Debugger) Run() error {
	fmt.Println("Running in debug mode.")

	for d.running {
		fmt.Println(d.bmo)
		fmt.Print("> ")

		command, args, err := d.parseInput()
		if err != nil {
			return err
		}

		d.execute(command, args)

		fmt.Println()
	}

	return nil
}

func (d *Debugger) parseInput() (string, []string, error) {
	input, err := d.reader.ReadString('\n')
	if err != nil {
		return "", nil, err
	}

	input = strings.TrimRight(input, "\n")
	args := strings.Split(input, " ")

	return args[0], args[1:], nil
}

func (d *Debugger) execute(command string, args []string) error {
	if command == "b" || command == "break" {
		if len(args) < 1 {
			fmt.Println("You must specify an address.")
			return nil
		}

		address, err := strconv.ParseInt(args[0], 16, 0)
		if err != nil {
			fmt.Println("Invalid address.")
			return nil
		}

		d.breaking = true
		d.breakpoint = uint16(address)
		fmt.Printf("Breakpoint set: %#04x.\n", d.breakpoint)
	} else if command == "c" || command == "clear" {
		d.breaking = false
		fmt.Println("Breakpoint cleared.")
	} else if command == "r" || command == "run" {
		if !d.breaking {
			fmt.Println("You must set a breakpoint first.")
			return nil
		}

		for {
			err := d.bmo.Step()
			if err != nil {
				return err
			}

			if d.bmo.GetPC() == d.breakpoint {
				break
			}
		}

		fmt.Println("Breakpoint reached.")
	} else if command == "" || command == "s" || command == "step" {
		err := d.bmo.Step()
		if err != nil {
			return err
		}
	} else if command == "q" || command == "quit" {
		d.running = false
		fmt.Print("Goodbye!")
	} else {
		fmt.Println("Unknown command.")
	}

	return nil
}
