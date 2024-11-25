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
