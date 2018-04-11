package interrupt

import (
	"github.com/bovarysme/bmo/mmu"
)

// Interrupt Controller registers' addresses
const (
	interruptEnableAddress  uint16 = 0xffff
	interruptRequestAddress uint16 = 0xff0f
)

// Interrupt Controller registers' masks
const (
	VBlank byte = 1 << iota
	LCDSTAT
	Timer
	Serial
	Joypad
)

type IC struct {
	mmu *mmu.MMU
}

func NewIC(mmu *mmu.MMU) *IC {
	return &IC{
		mmu: mmu,
	}
}

func (ic *IC) Check() (bool, int) {
	ie := ic.mmu.ReadByte(interruptEnableAddress)
	ir := ic.mmu.ReadByte(interruptRequestAddress)

	for i := 0; i <= 4; i++ {
		var mask byte = 1 << byte(i)

		enabled := ie&mask == mask
		requested := ir&mask == mask

		if enabled && requested {
			return true, i
		}
	}

	return false, 0
}

func (ic *IC) Clear(mask byte) {
	address := interruptRequestAddress

	ir := ic.mmu.ReadByte(address)
	ir &^= mask
	ic.mmu.WriteByte(address, ir)
}

func (ic *IC) Request(mask byte) {
	address := interruptRequestAddress

	ir := ic.mmu.ReadByte(address)
	ir |= mask
	ic.mmu.WriteByte(address, ir)
}
