package beemo

import (
	"github.com/bovarysme/bmo/cartridge"
	"github.com/bovarysme/bmo/cpu"
	"github.com/bovarysme/bmo/interrupt"
	"github.com/bovarysme/bmo/mmu"
	"github.com/bovarysme/bmo/ppu"
	"github.com/bovarysme/bmo/screen"
	"github.com/bovarysme/bmo/timer"
)

type BMO struct {
	cartridge mmu.Memory
	cpu       *cpu.CPU
	ic        *interrupt.IC
	mmu       *mmu.MMU
	ppu       *ppu.PPU
	screen    screen.Screen
	timer     *timer.Timer
}

func NewBMO(bootromPath, romPath string) (*BMO, error) {
	s, err := screen.NewSDLScreen()
	if err != nil {
		return nil, err
	}

	c, err := cartridge.NewCartridge(romPath)
	if err != nil {
		return nil, err
	}

	m, err := mmu.NewMMU(bootromPath, c)
	if err != nil {
		return nil, err
	}

	ic := interrupt.NewIC(m)

	p := ppu.NewPPU(m, ic)

	// XXX
	m.LinkPPU(p)

	return &BMO{
		cartridge: c,
		cpu:       cpu.NewCPU(m, ic),
		ic:        ic,
		mmu:       m,
		ppu:       p,
		screen:    s,
		timer:     timer.NewTimer(m, ic),
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

	b.timer.Step(cycles)

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
