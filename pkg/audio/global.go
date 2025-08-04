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
	audioEnabled bool

	// NR50
	vinLeft     byte
	leftVolume  byte
	vinRight    byte
	rightVolume byte
}

func (g *GlobalRegister) clear() {
	g.pulse1Left = false
	g.pulse2Left = false
	g.waveLeft = false
	g.noiseLeft = false
	g.pulse1Right = false
	g.pulse2Right = false
	g.waveRight = false
	g.noiseRight = false

	g.vinLeft = 0
	g.leftVolume = 0
	g.vinRight = 0
	g.rightVolume = 0
}
