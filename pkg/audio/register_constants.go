package audio

const (
	// Channel 1 (Pulse)
	nr10 = iota
	nr11
	nr12
	nr13
	nr14

	// Channel 2 (Pulse)
	nr21
	nr22
	nr23
	nr24

	// Channel 3 (Wave)
	nr30
	nr31
	nr32
	nr33
	nr34

	// Channel 4 (Noise)
	nr41
	nr42
	nr43
	nr44

	// Global
	nr50
	nr51
	nr52
)

var registerMasks = map[int]byte{
	nr10: 0x80,
	nr11: 0x3F,
	nr12: 0x00,
	nr13: 0xFF,
	nr14: 0xBF,
	nr21: 0x3F,
	nr22: 0x00,
	nr23: 0xFF,
	nr24: 0xBF,
	nr30: 0x7F,
	nr31: 0xFF,
	nr32: 0x9F,
	nr33: 0x00,
	nr34: 0xBF,
	nr41: 0xFF,
	nr42: 0x00,
	nr43: 0x00,
	nr44: 0xBF,
	nr50: 0x00,
	nr51: 0x00,
	nr52: 0x70,
}
