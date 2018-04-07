package apu

// Sound Channel 1 (Tone and Sweep)
const (
	NR10 uint16 = 0xff10 + iota // Sweep (R/W)
	NR11                        // Sound Length and Waveform Duty Cycle (R/W)
	NR12                        // Volume Envelope (R/W)
	NR13                        // Frequency lo (W)
	NR14                        // Frequency hi (R/W)
)

// Sound Channel 2 (Tone)
const (
	NR21 uint16 = 0xff16 + iota // Sound Length and Waveform Duty Cycle (R/W)
	NR22                        // Volume Envelope (R/W)
	NR23                        // Frequency lo (W)
	NR24                        // Frequency hi (R/W)
)

// Sound Channel 3 (Wave Output)
const (
	NR30 uint16 = 0xff1a + iota // Channel Enable (R/W)
	NR31                        // Sound Length (R/W)
	NR32                        // Output Level Selection (R/W)
	NR33                        // Frequency lo (W)
	NR34                        // Frequency hi (R/W)
)

// Sound Channel 4 (Noise)
const (
	NR41 uint16 = 0xff20 + iota // Sound Length (R/W)
	NR42                        // Volume Envelope (R/W)
	NR43                        // Polynomial Counter (R/W) ?
	NR44                        // Counter and consecutive; Initial (R/W) ?
)

// Sound Control Registers
const (
	NR50 uint16 = 0xff24 + iota // Left Right Enable and Output Level (R/W)
	NR51                        // Sound Output Terminal Selection (R/W)
	NR52                        // Sound Enable
)
