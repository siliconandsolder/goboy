package audio

type noiseRegister struct {
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

func (n *noiseRegister) cycleLengthTimer() bool {
	if n.lengthEnabled && n.lengthTimer > 0 {
		n.lengthTimer--
		if n.lengthTimer == 0 {
			return false
		}
	}

	return true
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
}

func (n *noiseRegister) getVolumeEnvelope() byte {
	var retVal byte = 0

	retVal |= n.volume << 4
	retVal |= n.envDirection << 3
	retVal |= n.envPace

	return retVal
}

func (n *noiseRegister) setFreqRandomness(value byte) {
	n.clockShift = value & 0xF >> 4
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

func (n *noiseRegister) setNoiseControl(value byte) bool {
	trigger := false
	if value>>7&1 == 1 {
		if n.lengthTimer == 0 {
			n.lengthTimer = LENGTH_TIMER_MAX - n.initLength
		}
		n.lfsr = 0x7FFF
		n.currentVolume = n.volume
		n.volumeTimer = n.envPace
		trigger = true
	}

	n.lengthEnabled = value>>6&1 == 1
	return trigger
}

func (n *noiseRegister) getNoiseControl() byte {
	if n.lengthEnabled {
		return 1 << 6
	} else {
		return 0
	}
}
