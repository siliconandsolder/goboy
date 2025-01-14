package audio

var waveVolume = [4]byte{
	4, 0, 1, 2,
}

type waveRegister struct {
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
	ram []byte

	freqTimer   uint16
	sampleIdx   byte
	lengthTimer uint16
}

func (w *waveRegister) cycleFrequencyTimer() {
	if w.freqTimer > 0 {
		w.freqTimer--
		if w.freqTimer == 0 {
			period := uint16(w.periodHigh)<<8 | uint16(w.periodLow)
			w.freqTimer = (2048 - period) * 2
			w.sampleIdx++
			w.sampleIdx &= 0x1F
		}
	}
}

func (w *waveRegister) cycleLengthTimer() bool {
	if w.lengthEnabled && w.lengthTimer > 0 {
		w.lengthTimer--
		if w.lengthTimer == 0 {
			return false
		}
	}

	return true
}

func (w *waveRegister) getSample() byte {
	var sample byte = 0
	if w.sampleIdx&1 == 0 {
		sample = w.ram[w.sampleIdx>>1] >> 4 // get the higher nibble of the sample
	} else {
		sample = w.ram[w.sampleIdx>>1] & 0xF // get the lower nibble of the sample
	}

	sample >>= waveVolume[w.output]
	return sample
}

func (w *waveRegister) getDAC() byte {
	var retVal byte

	if w.dacEnabled {
		retVal |= 1 << 7
	}

	return retVal
}

func (w *waveRegister) setPeriodHigh(value byte) bool {
	trigger := false
	w.lengthEnabled = (value>>6)&1 == 1
	w.periodHigh = value & 7

	if value>>7&1 == 1 {
		if w.lengthTimer == 0 {
			w.lengthTimer = LENGTH_TIMER_WAVE_MAX //- uint16(w.initLength)
		}
		period := uint16(w.periodHigh)<<8 | uint16(w.periodLow)
		w.freqTimer = (2048 - period) * 2
		w.sampleIdx = 0
		trigger = true
	}

	return trigger
}

func (w *waveRegister) getLengthEnabled() byte {
	if w.lengthEnabled {
		return 1 << 6
	} else {
		return 0
	}
}
