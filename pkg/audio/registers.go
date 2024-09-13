package audio

type GlobalRegister struct {
	Panning byte // NR51

	// NR52
	audioEnabled  bool
	pulse1Enabled bool
	pulse2Enabled bool
	waveEnabled   bool
	noiseEnabled  bool

	// NR50
	vinLeft     byte
	leftVolume  byte
	vinRight    byte
	rightVolume byte
}

type PulseRegister struct {
	// NR10 - Channel 1 only
	sweepPace      byte
	sweepDirection byte
	sweepStep      byte

	// NR11/NR21
	duty       byte
	initLength byte

	// NR12/NR22
	volume       byte
	envDirection byte
	envPace      byte

	// NR13/NR23
	periodLow byte

	// NR14/NR24
	periodHigh    byte
	lengthEnabled bool
}

type WaveRegister struct {
	// NR30
	dacEnabled bool

	// NR31
	initLength byte

	// NR32
	output byte

	// NR33
	periodLow byte

	// NR34
	periodHigh    byte
	lengthEnabled bool

	// NR35-NR40
	waveRam []byte
}

type NoiseRegister struct {
	// NR41
	initLength byte

	// NR42
	volume       byte
	envDirection byte
	envPace      byte

	// NR43
	clockShift   byte
	lfsrWidth    byte
	clockDivider byte

	// NR44
	lengthEnabled bool
}
