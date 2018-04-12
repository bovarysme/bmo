package beemo

import (
	"github.com/bovarysme/bmo/cpu"
	"github.com/bovarysme/bmo/interrupt"
	"github.com/bovarysme/bmo/mmu"
	"github.com/bovarysme/bmo/ppu"
	"github.com/bovarysme/bmo/screen"
)

type BMO struct {
	cpu    *cpu.CPU
	ic     *interrupt.IC
	mmu    *mmu.MMU
	ppu    *ppu.PPU
	screen screen.Screen
}

func NewBMO(rom []byte) (*BMO, error) {
	s, err := screen.NewSDLScreen()
	if err != nil {
		return nil, err
	}

	m := mmu.NewMMU(rom)
	ic := interrupt.NewIC(m)

	return &BMO{
		cpu:    cpu.NewCPU(m, ic),
		ic:     ic,
		mmu:    m,
		ppu:    ppu.NewPPU(m, ic),
		screen: s,
	}, nil
}

func (b *BMO) String() string {
	return b.cpu.String()
}

// XXX
func (b *BMO) GetPC() uint16 {
	return b.cpu.GetPC()
}

func (b *BMO) Step() error {
	cycles, err := b.cpu.Step()
	if err != nil {
		return err
	}

	b.ppu.Step(cycles)
	if b.ppu.VBlank {
		b.ppu.VBlank = false

		err = b.screen.Render(b.ppu.Pixels)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BMO) Run() error {
	for {
		err := b.Step()
		if err != nil {
			return err
		}
	}
}
