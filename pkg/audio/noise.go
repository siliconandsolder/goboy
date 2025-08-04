package audio

type noiseRegister struct {
	enabled bool

	// NR41
	initLength byte

	// NR42
	volume       byte
	envDirection byte
	envPace      byte
	dacEnabled   bool

	// NR43
	clockShift   byte
	lfsrWidth    byte
	clockDivider byte

	// NR44
	lengthEnabled bool

	freqTimer     byte
	lfsr          uint16
	lengthTimer   byte
	volumeTimer   byte
	currentVolume byte
}

func (n *noiseRegister) cycleFrequencyTimer() {
	if n.freqTimer == 0 {
		var divisor byte
		if n.clockDivider == 0 {
			divisor = 8
		} else {
			divisor = n.clockDivider << 4
		}
		n.freqTimer = divisor << n.clockShift

		xor := (n.lfsr & 1) ^ ((n.lfsr & 2) >> 1)
		n.lfsr = (n.lfsr >> 1) | (xor << 14)

		if n.lfsrWidth == 1 {
			n.lfsr &= ^uint16(1 << 6)
			n.lfsr |= xor << 6
		}
	} else {
		n.freqTimer--
	}
}

func (n *noiseRegister) cycleLengthTimer() {
	if n.lengthEnabled && n.lengthTimer > 0 {
		n.lengthTimer--
		if n.lengthTimer == 0 {
			n.enabled = false
		}
	}
}

func (n *noiseRegister) cycleVolumeTimer() {
	if n.volumeTimer > 0 {
		n.volumeTimer--
		if n.volumeTimer == 0 {
			n.volumeTimer = n.envPace

			if n.envDirection == 1 && n.currentVolume < 0xF {
				n.currentVolume++
			} else if n.envDirection == 0 && n.currentVolume > 0 {
				n.currentVolume--
			}
		}
	}
}

func (n *noiseRegister) getSample() byte {
	return byte(^n.lfsr&1) * n.currentVolume
}

func (n *noiseRegister) setVolumeEnvelope(value byte) {
	n.volume = value >> 4 & 0xF
	n.envDirection = value >> 3 & 1
	n.envPace = value & 7
	n.dacEnabled = value&0xF8 != 0
	if !n.dacEnabled {
		n.enabled = false
	}
}

func (n *noiseRegister) getVolumeEnvelope() byte {
	var retVal byte = 0

	retVal |= n.volume << 4
	retVal |= n.envDirection << 3
	retVal |= n.envPace

	return retVal
}

func (n *noiseRegister) setFreqRandomness(value byte) {
	n.clockShift = value >> 4 & 0xF
	n.lfsrWidth = value >> 3 & 1
	n.clockDivider = value & 7
}

func (n *noiseRegister) getFreqRandomness() byte {
	var retVal byte = 0

	retVal |= n.clockShift << 4
	retVal |= n.lfsrWidth << 3
	retVal |= n.clockDivider

	return retVal
}

func (n *noiseRegister) setNoiseControl(value byte) {
	if value>>7&1 == 1 {
		if n.lengthTimer == 0 {
			n.lengthTimer = LENGTH_TIMER_MAX
		}
		n.lfsr = 0x7FFF
		n.currentVolume = n.volume
		n.volumeTimer = n.envPace
		if n.dacEnabled {
			n.enabled = true
		}
	}

	n.lengthEnabled = value>>6&1 == 1
}

func (n *noiseRegister) getNoiseControl() byte {
	if n.lengthEnabled {
		return 1 << 6
	} else {
		return 0
	}
}

func (n *noiseRegister) clear() {
	n.initLength = 0
	n.volume = 0
	n.envDirection = 0
	n.envPace = 0
	n.clockShift = 0
	n.lfsrWidth = 0
	n.clockDivider = 0
	n.lengthEnabled = false
	n.freqTimer = 0
	n.lfsr = 0
	n.lengthTimer = 0
	n.volumeTimer = 0
	n.currentVolume = 0
}
