package audio

var pulseDutyTable = [4][8]byte{
	{0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 1, 1, 1},
	{0, 1, 1, 1, 1, 1, 1, 0},
}

type pulseRegister struct {
	enabled bool

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

	isChannel1      bool
	freqTimer       uint16
	wavePos         byte
	lengthTimer     byte
	sweepTimer      byte
	sweepEnabled    bool
	shadowSweep     uint16
	sweepCalculated bool
	volumeTimer     byte
	currentVolume   byte
	envEnabled      bool
	dacEnabled      bool
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

func (p *pulseRegister) cycleLengthTimer() {
	if p.lengthEnabled && p.lengthTimer > 0 {
		p.lengthTimer--
		if p.lengthTimer == 0 {
			p.enabled = false
		}
	}
}

func (p *pulseRegister) cycleSweepTimer() {
	if p.sweepTimer > 0 {
		p.sweepTimer--
	}
	if p.sweepTimer == 0 {
		p.sweepTimer = p.sweepPace
		if p.sweepTimer == 0 {
			p.sweepTimer = 8 // delay for 8 iterations before checking again
		}

		if p.sweepEnabled && p.sweepPace > 0 {
			newPeriod := p.pulse1IterateSweep()

			if newPeriod <= 0x7FF && p.sweepStep > 0 {
				p.periodLow = byte(newPeriod & 0xFF)
				p.periodHigh = byte((newPeriod >> 8) & 7)
				p.shadowSweep = newPeriod

				p.pulse1IterateSweep() // overflow check
			}
		}
	}
}

func (p *pulseRegister) pulse1IterateSweep() uint16 {
	newPeriod := p.shadowSweep >> p.sweepStep

	if p.sweepDirection == 0 {
		newPeriod += p.shadowSweep
	} else {
		newPeriod = p.shadowSweep - newPeriod
	}

	if newPeriod > 0x7FF {
		p.enabled = false
	}

	// Clearing the sweep negate mode bit in NR10 after at least one sweep calculation has been made using
	// the negate mode since the last trigger causes the channel to be immediately disabled
	if p.sweepDirection == 1 {
		p.sweepCalculated = true
	}

	return newPeriod
}

func (p *pulseRegister) cycleVolumeTimer() {
	if p.volumeTimer > 0 {
		p.volumeTimer--
		if p.volumeTimer == 0 {
			p.volumeTimer = p.envPace

			if p.envEnabled && p.envPace > 0 {
				if p.envDirection == 1 && p.currentVolume < 0xF {
					p.currentVolume++
				} else if p.envDirection == 0 && p.currentVolume > 0 {
					p.currentVolume--
				}
			}

			if p.currentVolume == 0 || p.currentVolume == 15 {
				p.envEnabled = false
			}

		}
	}
}

func (p *pulseRegister) getSample() byte {
	return pulseDutyTable[p.duty][p.wavePos] * p.currentVolume
}

func (p *pulseRegister) setSweep(value byte) {
	if p.sweepDirection == 1 && (value>>3)&1 == 0 && p.sweepCalculated {
		p.enabled = false
	}

	p.sweepPace = (value >> 4) & 7
	p.sweepDirection = (value >> 3) & 1
	p.sweepStep = value & 7
}

func (p *pulseRegister) setLengthDuty(value byte) {
	p.duty = (value >> 6) & 3
	p.initLength = value & 0x3F
	p.lengthTimer = LENGTH_TIMER_MAX - p.initLength
}

func (p *pulseRegister) setVolumeEnvelope(value byte) {
	p.volume = (value >> 4) & 0xF
	p.envDirection = (value >> 3) & 1
	p.envPace = value & 7
	p.dacEnabled = value&0xF8 != 0
	if !p.dacEnabled {
		p.enabled = false
	}
}

func (p *pulseRegister) setPeriodLow(value byte) {
	p.periodLow = value
}

func (p *pulseRegister) setPeriodHigh(value byte, isCycleLengthTimerStep bool) {
	p.periodHigh = value & 7

	if value&0x80 == 0x80 {
		if p.dacEnabled {
			p.enabled = true
		}
		if p.lengthTimer == 0 {
			if p.lengthEnabled && !isCycleLengthTimerStep {
				p.lengthTimer = LENGTH_TIMER_MAX - 1
			} else {
				p.lengthTimer = LENGTH_TIMER_MAX
			}
		}
		p.currentVolume = p.volume
		p.volumeTimer = p.envPace

		period := uint16(p.periodHigh&7)<<8 | uint16(p.periodLow)
		p.freqTimer = (2048 - period) * 4

		p.sweepCalculated = false
		p.sweepTimer = p.sweepPace
		if p.sweepTimer == 0 {
			p.sweepTimer = 8
		}
		p.shadowSweep = period
		p.sweepEnabled = p.sweepPace > 0 || p.sweepStep > 0
		if p.sweepStep > 0 {
			p.pulse1IterateSweep()
		}
		p.envEnabled = true
	}

	if value&0x40 == 0x40 && !p.lengthEnabled && !isCycleLengthTimerStep {
		if p.lengthTimer > 0 {
			p.lengthTimer--
			if p.lengthTimer == 0 {
				p.enabled = false
			}
		}
	}
	p.lengthEnabled = value&0x40 == 0x40
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

func (p *pulseRegister) clear() {
	p.sweepPace = 0
	p.sweepDirection = 0
	p.sweepStep = 0
	p.duty = 0
	p.initLength = 0
	p.volume = 0
	p.envDirection = 0
	p.envPace = 0
	p.periodLow = 0
	p.periodHigh = 0
	p.lengthEnabled = false
	p.freqTimer = 0
	p.wavePos = 0
	p.lengthTimer = 0
	p.sweepTimer = 0
	p.sweepEnabled = false
	p.shadowSweep = 0
	p.volumeTimer = 0
	p.currentVolume = 0
	p.envEnabled = false
	p.dacEnabled = false
}
