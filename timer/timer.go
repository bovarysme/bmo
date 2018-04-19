package timer

import (
	"github.com/bovarysme/bmo/interrupt"
	"github.com/bovarysme/bmo/mmu"
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
	ic  *interrupt.IC
	mmu *mmu.MMU

	cycles int
}

func NewTimer(m *mmu.MMU, ic *interrupt.IC) *Timer {
	return &Timer{
		ic:  ic,
		mmu: m,
	}
}

func (t *Timer) Step(cycles int) {
	control := t.mmu.ReadByte(TAC)

	enabled := control&TimerStart == TimerStart
	if !enabled {
		return
	}

	inputClock := control & InputClockSelect
	freq := freqs[inputClock]

	t.cycles += cycles
	if t.cycles >= freq {
		t.cycles -= freq

		counter := t.mmu.ReadByte(TIMA)
		counter++
		if counter == 0 {
			modulo := t.mmu.ReadByte(TMA)
			counter = modulo

			t.ic.Request(interrupt.Timer)
		}

		t.mmu.WriteByte(TIMA, counter)
	}
}
