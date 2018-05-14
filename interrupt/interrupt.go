package interrupt

// Interrupt Controller registers' addresses
const (
	IR uint16 = 0xff0f // Interrupt Request
	IE uint16 = 0xffff // Interrupt Enable
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
	ir byte
	ie byte
}

func NewIC() *IC {
	return &IC{}
}

func (ic *IC) ReadByte(address uint16) byte {
	var value byte

	switch address {
	case IR:
		value = ic.ir
	case IE:
		value = ic.ie
	}

	return value
}

func (ic *IC) WriteByte(address uint16, value byte) {
	switch address {
	case IR:
		ic.ir = value
	case IE:
		ic.ie = value
	}
}

func (ic *IC) Check() (bool, int) {
	for i := 0; i < 5; i++ {
		var mask byte = 1 << byte(i)

		enabled := ic.ie&mask == mask
		requested := ic.ir&mask == mask

		if enabled && requested {
			return true, i
		}
	}

	return false, 0
}

func (ic *IC) Clear(mask byte) {
	ic.ir &^= mask
}

func (ic *IC) Request(mask byte) {
	ic.ir |= mask
}
