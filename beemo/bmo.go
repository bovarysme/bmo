package beemo

import (
	"github.com/bovarysme/bmo/cartridge"
	"github.com/bovarysme/bmo/cpu"
	"github.com/bovarysme/bmo/input"
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
	joypad    *input.Joypad
	mmu       *mmu.MMU
	ppu       *ppu.PPU
	timer     *timer.Timer

	keys   input.Keys
	screen screen.Screen

	running bool
}

func NewBMO(bootromPath, romPath string, screenScale int) (*BMO, error) {
	c, err := cartridge.NewCartridge(romPath)
	if err != nil {
		return nil, err
	}

	m, err := mmu.NewMMU(bootromPath, c)
	if err != nil {
		return nil, err
	}

	ic := interrupt.NewIC()

	joypad := input.NewJoypad(ic)
	p := ppu.NewPPU(m, ic)

	// XXX
	m.LinkIC(ic)
	m.LinkJoypad(joypad)
	m.LinkPPU(p)

	keys := input.NewSDLKeys(joypad)

	s, err := screen.NewSDLScreen(screenScale)
	if err != nil {
		return nil, err
	}

	return &BMO{
		cartridge: c,
		cpu:       cpu.NewCPU(m, ic),
		ic:        ic,
		joypad:    joypad,
		mmu:       m,
		ppu:       p,
		timer:     timer.NewTimer(m, ic),

		keys:   keys,
		screen: s,

		running: true,
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

		event := b.keys.Read()
		if event == input.Quit {
			b.running = false
		}
	}

	b.timer.Step(cycles)

	return nil
}

func (b *BMO) Run() error {
	for b.running {
		err := b.Step()
		if err != nil {
			return err
		}
	}

	b.screen.Shutdown()

	return nil
}
