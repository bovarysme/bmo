package beemo

import (
	"github.com/bovarysme/bmo/cartridge"
	"github.com/bovarysme/bmo/cpu"
	"github.com/bovarysme/bmo/interrupt"
	"github.com/bovarysme/bmo/joypad"
	"github.com/bovarysme/bmo/mmu"
	"github.com/bovarysme/bmo/ppu"
	"github.com/bovarysme/bmo/screen"
	"github.com/bovarysme/bmo/timer"
)

type BMO struct {
	cartridge mmu.Memory
	cpu       *cpu.CPU
	ic        *interrupt.IC
	joypad    *joypad.Joypad
	mmu       *mmu.MMU
	ppu       *ppu.PPU
	timer     *timer.Timer

	screen screen.Screen
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

	j := joypad.NewJoypad(ic)
	p := ppu.NewPPU(m, ic)

	// XXX
	m.LinkPPU(p)
	m.LinkJoypad(j)

	return &BMO{
		cartridge: c,
		cpu:       cpu.NewCPU(m, ic),
		ic:        ic,
		joypad:    j,
		mmu:       m,
		ppu:       p,
		timer:     timer.NewTimer(m, ic),

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

		joypad.Step(b.joypad)
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
