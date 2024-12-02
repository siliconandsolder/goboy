package audio

var pulseDutyTable = [4][8]byte{
	{0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 1, 1, 1},
	{0, 1, 1, 1, 1, 1, 1, 0},
}

type pulseRegister struct {
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

	isChannel1    bool
	freqTimer     uint16
	wavePos       byte
	lengthTimer   byte
	sweepTimer    byte
	sweepEnabled  bool
	shadowSweep   uint16
	volumeTimer   byte
	currentVolume byte
}

func (p *pulseRegister) cycleFrequencyTimer() {
	if p.freqTimer > 0 {
		p.freqTimer--
		if p.freqTimer == 0 {
			// reload timer with period value
			period := uint16(p.periodHigh)<<8 | uint16(p.periodLow)
			p.freqTimer = (2048 - period) * 4
			// update wave duty position
			p.wavePos++
			p.wavePos &= 7
		}
	}
}

func (p *pulseRegister) cycleLengthTimer() bool {
	if p.lengthEnabled && p.lengthTimer > 0 {
		p.lengthTimer--
		if p.lengthTimer == 0 {
			return false
		}
	}

	return true
}

func (p *pulseRegister) cycleSweepTimer() bool {
	if p.sweepTimer > 0 {
		p.sweepTimer--
	}
	if p.sweepTimer == 0 {
		p.sweepTimer = p.sweepPace
		if p.sweepTimer == 0 {
			p.sweepTimer = 8 // delay for 8 iterations before checking again
		}

		if p.sweepEnabled && p.sweepPace > 0 {
			newPeriod, channelEnabled := p.pulse1IterateSweep()

			if newPeriod <= 0x7FF && p.sweepStep > 0 {
				p.periodLow = byte(newPeriod & 0xFF)
				p.periodHigh = byte((newPeriod >> 8) & 7)
				p.shadowSweep = uint16(p.periodHigh&7)<<8 | uint16(p.periodLow)

				p.pulse1IterateSweep() // overflow check
			}
			p.pulse1IterateSweep()

			return channelEnabled
		}
	}

	return true
}

func (p *pulseRegister) pulse1IterateSweep() (uint16, bool) {
	newPeriod := p.shadowSweep >> p.sweepStep

	if p.sweepDirection == 0 {
		newPeriod += p.shadowSweep
	} else {
		newPeriod = p.shadowSweep - newPeriod
	}

	if newPeriod > 0x7FF {
		return newPeriod, false
	}

	return newPeriod, true
}

func (p *pulseRegister) cycleVolumeTimer() {
	if p.volumeTimer > 0 {
		p.volumeTimer--
		if p.volumeTimer == 0 {
			p.volumeTimer = p.envPace

			if p.envDirection == 1 && p.currentVolume < 0xF {
				p.currentVolume++
			} else if p.envDirection == 0 && p.currentVolume > 0 {
				p.currentVolume--
			}
		}
	}
}

func (p *pulseRegister) getSample() byte {
	return pulseDutyTable[p.duty][p.wavePos] * p.currentVolume
}

func (p *pulseRegister) setSweep(value byte) {
	p.sweepPace = (value >> 4) & 7
	p.sweepDirection = (value >> 3) & 1
	p.sweepStep = value & 7
}

func (p *pulseRegister) setLengthDuty(value byte) {
	p.duty = (value >> 6) & 3
	p.initLength = value & 0x3F
}

func (p *pulseRegister) setVolumeEnvelope(value byte) {
	p.volume = (value >> 4) & 0xF
	p.envDirection = (value >> 3) & 1
	p.envPace = value & 7
}

func (p *pulseRegister) setPeriodLow(value byte) {
	p.periodLow = value
}

func (p *pulseRegister) setPeriodHigh(value byte) bool {
	p.lengthEnabled = (value>>6)&1 == 1
	p.periodHigh = value & 7

	if value&0x80 == 0x80 {
		if p.lengthTimer == 0 {
			p.lengthTimer = LENGTH_TIMER_MAX - p.initLength
		}
		p.currentVolume = p.volume
		p.volumeTimer = p.envPace

		period := uint16(p.periodHigh&7)<<8 | uint16(p.periodLow)
		p.freqTimer = (2048 - period) * 4

		p.sweepTimer = p.sweepPace
		if p.sweepTimer == 0 {
			p.sweepTimer = 8
		}
		p.shadowSweep = period
		p.sweepEnabled = p.sweepTimer > 0 || p.sweepStep > 0
		return true
	}

	return false
}

func (p *pulseRegister) getSweep() byte {
	var retVal byte = 0

	retVal |= p.sweepPace << 4
	retVal |= p.sweepDirection << 3
	retVal |= p.sweepStep

	return retVal
}

func (p *pulseRegister) getLengthDuty() byte {
	var retVal byte = 0

	retVal |= p.duty << 6
	retVal |= p.initLength

	return retVal
}

func (p *pulseRegister) getVolumeEnvelope() byte {
	var retVal byte = 0

	retVal |= p.volume << 4
	retVal |= p.envDirection << 3
	retVal |= p.envPace

	return retVal
}

func (p *pulseRegister) getPeriodHigh() byte {
	var retVal byte = 0

	if p.lengthEnabled {
		retVal |= 1 << 6
	}

	return retVal
}
