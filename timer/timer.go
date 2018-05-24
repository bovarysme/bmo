package timer

import (
	"github.com/bovarysme/bmo/interrupt"
)

const (
	DIV  uint16 = 0xff04 + iota // Divider Register (R/W)
	TIMA                        // Timer Counter (R/W)
	TMA                         // Timer Modulo (R/W)
	TAC                         // Timer Control (R/W)
)

// Timer Control's masks
const (
	InputClockSelect byte = 0x3
	TimerStart       byte = 1 << 2
)

var freqs = [4]int{256, 4, 16, 64}

type Timer struct {
	ic *interrupt.IC

	div  byte
	tima byte
	tma  byte
	tac  byte

	cycles int
}

func NewTimer(ic *interrupt.IC) *Timer {
	return &Timer{
		ic: ic,
	}
}

func (t *Timer) ReadByte(address uint16) byte {
	var value byte

	switch address {
	case DIV:
		value = t.div
	case TIMA:
		value = t.tima
	case TMA:
		value = t.tma
	case TAC:
		value = t.tac
	}

	return value
}

func (t *Timer) WriteByte(address uint16, value byte) {
	switch address {
	case DIV:
		t.div = value
	case TIMA:
		t.tima = value
	case TMA:
		t.tma = value
	case TAC:
		t.tac = value
	}
}

func (t *Timer) Step(cycles int) {
	enabled := t.tac&TimerStart == TimerStart
	if !enabled {
		return
	}

	inputClock := t.tac & InputClockSelect
	freq := freqs[inputClock]

	t.cycles += cycles
	if t.cycles >= freq {
		t.cycles -= freq

		t.tima++
		if t.tima == 0 {
			t.tima = t.tma

			t.ic.Request(interrupt.Timer)
		}
	}
}
