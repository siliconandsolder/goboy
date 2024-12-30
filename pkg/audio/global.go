package audio

type GlobalRegister struct {
	// NR51
	pulse1Left  bool
	pulse2Left  bool
	waveLeft    bool
	noiseLeft   bool
	pulse1Right bool
	pulse2Right bool
	waveRight   bool
	noiseRight  bool

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
