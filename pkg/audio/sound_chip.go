package audio

import (
	"github.com/veandco/go-sdl2/sdl"
)

const LENGTH_TIMER_MAX = 64
const LENGTH_TIMER_WAVE_MAX = 256
const CYCLES_PER_SAMPLE = 95

var pulseDutyTable = [4][8]byte{
	{0, 0, 0, 0, 0, 0, 0, 1},
	{0, 0, 0, 0, 0, 0, 1, 1},
	{0, 0, 0, 0, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 0, 0},
}

var waveVolume = [4]byte{
	4, 0, 1, 2,
}

type SoundChip struct {
	Global GlobalRegister
	Pulse1 PulseRegister
	Pulse2 PulseRegister
	Wave   WaveRegister
	Noise  NoiseRegister

	pulse1SweepTimer   byte
	pulse1SweepEnabled bool
	pulse1ShadowSweep  uint16
	pulse1VolumeTimer  byte
	pulse1Volume       byte
	pulse1LengthTimer  byte
	pulse1FreqTimer    uint16
	pulse1WavePos      byte

	pulse2VolumeTimer byte
	pulse2Volume      byte
	pulse2LengthTimer byte
	pulse2FreqTimer   uint16
	pulse2WavePos     byte

	waveSampleIdx   byte
	waveNibbleLow   bool
	waveFreqTimer   uint16
	waveLengthTimer uint16

	noiseLengthTimer byte
	noiseFreqTimer   byte
	noiseLfsr        uint16
	noiseVolumeTimer byte
	noiseVolume      byte

	frameSequencer byte
	cyclesToSample byte
	player         *Player
}

func NewSoundChip(p *Player) *SoundChip {
	return &SoundChip{
		Global: GlobalRegister{},
		Pulse1: PulseRegister{},
		Pulse2: PulseRegister{},
		Wave: WaveRegister{
			waveRam: make([]byte, 0x20),
		},
		Noise:             NoiseRegister{},
		pulse1SweepTimer:  0,
		pulse1VolumeTimer: 0,
		pulse1Volume:      0,
		pulse1LengthTimer: 0,
		pulse1FreqTimer:   0,
		pulse1WavePos:     0,
		pulse2VolumeTimer: 0,
		pulse2Volume:      0,
		pulse2LengthTimer: 0,
		pulse2FreqTimer:   0,
		pulse2WavePos:     0,
		waveSampleIdx:     0,
		waveNibbleLow:     false,
		waveFreqTimer:     0,
		waveLengthTimer:   0,
		noiseLengthTimer:  0,
		noiseFreqTimer:    0,
		noiseLfsr:         0,
		noiseVolumeTimer:  0,
		noiseVolume:       0,
		frameSequencer:    0,
		cyclesToSample:    CYCLES_PER_SAMPLE,
		player:            p,
	}
}

func (s *SoundChip) Cycle(cycles byte) {
	for i := byte(0); i < cycles; i++ {

		if s.pulse1FreqTimer > 0 {
			s.pulse1FreqTimer--
			if s.pulse1FreqTimer == 0 {
				// reload timer with period value
				period := uint16(s.Pulse1.periodHigh)<<8 | uint16(s.Pulse1.periodLow)
				s.pulse1FreqTimer = (2048 - period) * 4
				// update wave duty position
				s.pulse1WavePos++
				s.pulse1WavePos &= 7
			}
		}

		if s.pulse2FreqTimer > 0 {
			s.pulse2FreqTimer--
			if s.pulse2FreqTimer == 0 {
				// reload timer with period value
				period := uint16(s.Pulse2.periodHigh)<<8 | uint16(s.Pulse2.periodLow)
				s.pulse2FreqTimer = (2048 - period) * 4
				// update wave duty position
				s.pulse2WavePos++
				s.pulse2WavePos &= 7
			}
		}

		if s.waveFreqTimer > 0 {
			s.waveFreqTimer--
			if s.waveFreqTimer == 0 {
				period := uint16(s.Wave.periodHigh)<<8 | uint16(s.Wave.periodLow)
				s.waveFreqTimer = (2048 - period) * 2
				s.waveSampleIdx++
				s.waveSampleIdx &= 0x1F
			}
		}

		if s.noiseFreqTimer > 0 {
			s.noiseFreqTimer--
			if s.noiseFreqTimer == 0 {
				// TODO: Fix divisor
				var divisor byte
				if s.Noise.clockDivider == 0 {
					divisor = 8
				} else {
					divisor = s.Noise.clockDivider << 4
				}
				s.noiseFreqTimer = divisor << s.Noise.clockShift

				xor := (s.noiseLfsr & 1) ^ ((s.noiseLfsr & 2) >> 1)
				s.noiseLfsr = (s.noiseLfsr >> 1) | (xor << 14)

				if s.Noise.lfsrWidth == 1 {
					s.noiseLfsr &= ^uint16(1 << 6)
					s.noiseLfsr |= xor << 6
				}
			}
		}

		s.cyclesToSample--
		if s.cyclesToSample == 0 {
			s.cyclesToSample = CYCLES_PER_SAMPLE

			var pulse1Sample byte = 0
			var pulse2Sample byte = 0
			var waveSample byte = 0
			var noiseSample byte = 0

			//if s.Global.pulse1Enabled {
			//	pulse1Sample = pulseDutyTable[s.Pulse1.duty][s.pulse1WavePos] * s.pulse1Volume
			//}
			//
			//if s.Global.pulse2Enabled {
			//	pulse2Sample = pulseDutyTable[s.Pulse2.duty][s.pulse2WavePos] * s.pulse2Volume
			//}

			if s.Global.waveEnabled {
				if s.waveSampleIdx&1 == 0 {
					waveSample = s.Wave.waveRam[s.waveSampleIdx>>1] >> 4 // get the higher nibble of the sample
				} else {
					waveSample = s.Wave.waveRam[s.waveSampleIdx>>1] & 0xF // get the lower nibble of the sample
				}

				waveSample >>= waveVolume[s.Wave.output]
			}

			//
			//if s.Global.noiseEnabled {
			//	noiseSample = byte(^s.noiseLfsr&1) * s.noiseVolume
			//}

			mixedSample := float32(pulse1Sample+pulse2Sample+waveSample+noiseSample/4.0) / 50.0
			s.player.SendSample(mixedSample)
			for len(s.player.channel) > 2048 {
				sdl.Delay(1)
			}
		}
	}
}

func (s *SoundChip) CycleFrameSequencer() {
	if s.frameSequencer%2 == 0 {
		if s.Pulse1.lengthEnabled && s.pulse1LengthTimer > 0 {
			s.pulse1LengthTimer--
			if s.pulse1LengthTimer == 0 {
				s.Global.pulse1Enabled = false // disable channel 1
			}
		}

		if s.Pulse2.lengthEnabled && s.pulse2LengthTimer > 0 {
			s.pulse2LengthTimer--
			if s.pulse2LengthTimer == 0 {
				s.Global.pulse2Enabled = false // disable channel 2
			}
		}

		if s.Wave.lengthEnabled && s.waveLengthTimer > 0 {
			s.waveLengthTimer--
			if s.waveLengthTimer == 0 {
				s.Global.waveEnabled = false // disable channel 3
			}
		}

		if s.Noise.lengthEnabled && s.noiseLengthTimer > 0 {
			s.noiseLengthTimer--
			if s.noiseLengthTimer == 0 {
				s.Global.noiseEnabled = false // disable channel 2
			}
		}
	}

	if s.frameSequencer == 2 || s.frameSequencer == 6 {
		if s.pulse1SweepTimer > 0 {
			s.pulse1SweepTimer--
		}
		if s.pulse1SweepTimer == 0 {
			s.pulse1SweepTimer = s.Pulse1.sweepPace
			if s.pulse1SweepTimer == 0 {
				s.pulse1SweepTimer = 8 // delay for 8 iterations before checking again
			}

			if s.pulse1SweepEnabled && s.Pulse1.sweepPace > 0 {
				newPeriod := s.pulse1IterateSweep()

				if newPeriod <= 0x7FF && s.Pulse1.sweepStep > 0 {
					s.Pulse1.periodLow = byte(newPeriod & 0xFF)
					s.Pulse1.periodHigh = byte((newPeriod >> 8) & 7)
					s.pulse1ShadowSweep = uint16(s.Pulse1.periodHigh&7)<<8 | uint16(s.Pulse1.periodLow)

					s.pulse1IterateSweep() // overflow check
				}
				s.pulse1IterateSweep()
			}
		}
	}

	if s.frameSequencer == 7 {
		// cycle volume
		if s.pulse1VolumeTimer > 0 {
			s.pulse1VolumeTimer--
			if s.pulse1VolumeTimer == 0 {
				s.pulse1VolumeTimer = s.Pulse1.envPace

				if s.Pulse1.envDirection == 1 && s.pulse1Volume < 0xF {
					s.pulse1Volume++
				} else if s.Pulse1.envDirection == 0 && s.pulse1Volume > 0 {
					s.pulse1Volume--
				}
			}
		}

		if s.pulse2VolumeTimer > 0 {
			s.pulse2VolumeTimer--
			if s.pulse2VolumeTimer == 0 {
				s.pulse2VolumeTimer = s.Pulse2.envPace

				if s.Pulse2.envDirection == 1 && s.pulse2Volume < 0xF {
					s.pulse2Volume++
				} else if s.Pulse2.envDirection == 0 && s.pulse2Volume > 0 {
					s.pulse2Volume--
				}
			}
		}

		if s.noiseVolumeTimer > 0 {
			s.noiseVolumeTimer--
			if s.noiseVolumeTimer == 0 {
				s.noiseVolumeTimer = s.Noise.envPace

				if s.Noise.envDirection == 1 && s.noiseVolume < 0xF {
					s.noiseVolume++
				} else if s.Noise.envDirection == 0 && s.noiseVolume > 0 {
					s.noiseVolume--
				}
			}
		}
	}

	s.frameSequencer++
	s.frameSequencer &= 7
}

func (s *SoundChip) pulse1IterateSweep() uint16 {
	newPeriod := s.pulse1ShadowSweep >> s.Pulse1.sweepStep

	if s.Pulse1.sweepDirection == 0 {
		newPeriod += s.pulse1ShadowSweep
	} else {
		newPeriod = s.pulse1ShadowSweep - newPeriod
	}

	if newPeriod > 0x7FF {
		s.Global.pulse1Enabled = false // turn off channel 1
	}

	return newPeriod
}

func (s *SoundChip) SetMasterControl(value byte) {
	s.Global.audioEnabled = value>>7&1 == 1
	s.Global.noiseEnabled = value>>3&1 == 1
	s.Global.waveEnabled = value>>2&1 == 1
	s.Global.pulse2Enabled = value>>1&1 == 1
	s.Global.pulse1Enabled = value&1 == 1
}

func (s *SoundChip) GetMasterControl() byte {
	var retVal byte = 0
	if s.Global.audioEnabled {
		retVal |= 1 << 7
	}

	if s.Global.noiseEnabled {
		retVal |= 1 << 3
	}

	if s.Global.waveEnabled {
		retVal |= 1 << 2
	}

	if s.Global.pulse2Enabled {
		retVal |= 1 << 1
	}

	if s.Global.pulse1Enabled {
		retVal |= 1
	}

	return retVal
}

func (s *SoundChip) SetMasterVolume(value byte) {
	s.Global.vinLeft = value >> 7 & 1
	s.Global.leftVolume = value >> 4 & 7
	s.Global.vinRight = value >> 3 & 1
	s.Global.leftVolume = value & 7
}

func (s *SoundChip) GetMasterVolume() byte {
	var retVal byte = 0

	retVal |= s.Global.vinLeft << 7
	retVal |= s.Global.leftVolume << 4
	retVal |= s.Global.vinRight << 3
	retVal |= s.Global.rightVolume

	return retVal
}

func (s *SoundChip) SetPulse1Sweep(value byte) {
	s.Pulse1.sweepPace = (value >> 4) & 7
	s.Pulse1.sweepDirection = (value >> 3) & 1
	s.Pulse1.sweepStep = value & 7
}

func (s *SoundChip) GetPulse1Sweep() byte {
	var retVal byte = 0

	retVal |= s.Pulse1.sweepPace << 4
	retVal |= s.Pulse1.sweepDirection << 3
	retVal |= s.Pulse1.sweepStep

	return retVal
}

func (s *SoundChip) SetPulse1LengthDuty(value byte) {
	s.Pulse1.duty = (value >> 6) & 3
	s.Pulse1.initLength = value & 0x3F
}

func (s *SoundChip) GetPulse1LengthDuty() byte {
	var retVal byte = 0

	retVal |= s.Pulse1.duty << 6
	retVal |= s.Pulse1.initLength

	return retVal
}

func (s *SoundChip) SetPulse1VolumeEnvelope(value byte) {
	s.Pulse1.volume = (value >> 4) & 0xF
	s.Pulse1.envDirection = (value >> 3) & 1
	s.Pulse1.envPace = value & 7
}

func (s *SoundChip) GetPulse1VolumeEnvelope() byte {
	var retVal byte = 0

	retVal |= s.Pulse1.volume << 4
	retVal |= s.Pulse1.envDirection << 3
	retVal |= s.Pulse1.envPace

	return retVal
}

func (s *SoundChip) SetPulse1PeriodLow(value byte) {
	s.Pulse1.periodLow = value
}

func (s *SoundChip) SetPulse1PeriodHigh(value byte) {
	s.Pulse1.lengthEnabled = (value>>6)&1 == 1
	s.Pulse1.periodHigh = value & 7

	if value&0x80 == 0x80 {
		if s.pulse1LengthTimer == 0 {
			s.pulse1LengthTimer = LENGTH_TIMER_MAX - s.Pulse1.initLength
		}
		s.pulse1Volume = s.Pulse1.volume
		s.pulse1VolumeTimer = s.Pulse1.envPace

		period := uint16(s.Pulse1.periodHigh&7)<<8 | uint16(s.Pulse1.periodLow)
		s.pulse1FreqTimer = (2048 - period) * 4

		s.pulse1SweepTimer = s.Pulse1.sweepPace
		if s.pulse1SweepTimer == 0 {
			s.pulse1SweepTimer = 8
		}
		s.pulse1ShadowSweep = period
		s.pulse1SweepEnabled = s.pulse1SweepTimer > 0 || s.Pulse1.sweepStep > 0
		if s.Pulse1.sweepStep > 0 {
			s.pulse1IterateSweep()
		}
		s.Global.pulse1Enabled = true
	}
}

func (s *SoundChip) GetPulse1PeriodHigh() byte {
	var retVal byte = 0

	if s.Pulse1.lengthEnabled {
		retVal |= 1 << 6
	}

	return retVal
}

func (s *SoundChip) SetPulse2LengthDuty(value byte) {
	s.Pulse2.duty = (value >> 6) & 3
	s.Pulse2.initLength = value & 0x3F
}

func (s *SoundChip) GetPulse2LengthDuty() byte {
	var retVal byte = 0

	retVal |= s.Pulse2.duty << 6
	retVal |= s.Pulse2.initLength

	return retVal
}

func (s *SoundChip) SetPulse2VolumeEnvelope(value byte) {
	s.Pulse2.volume = (value >> 4) & 0xF
	s.Pulse2.envDirection = (value >> 3) & 1
	s.Pulse2.envPace = value & 7
}

func (s *SoundChip) GetPulse2VolumeEnvelope() byte {
	var retVal byte = 0

	retVal |= s.Pulse2.volume << 4
	retVal |= s.Pulse2.envDirection << 3
	retVal |= s.Pulse2.envPace

	return retVal
}

func (s *SoundChip) SetPulse2PeriodLow(value byte) {
	s.Pulse2.periodLow = value
}

func (s *SoundChip) SetPulse2PeriodHigh(value byte) {
	s.Pulse2.lengthEnabled = (value>>6)&1 == 1
	s.Pulse2.periodHigh = value & 7

	if value>>7 == 1 {
		if s.pulse2LengthTimer == 0 {
			s.pulse2LengthTimer = LENGTH_TIMER_MAX - s.Pulse1.initLength
		}
		s.pulse2Volume = s.Pulse2.volume
		s.pulse2VolumeTimer = s.Pulse2.envPace

		period := uint16(s.Pulse2.periodHigh&7)<<8 | uint16(s.Pulse2.periodLow)
		s.pulse2FreqTimer = (2048 - period) * 4
		s.Global.pulse2Enabled = true
	}
}

func (s *SoundChip) GetPulse2PeriodHigh() byte {
	var retVal byte = 0

	if s.Pulse2.lengthEnabled {
		retVal |= 1 << 6
	}

	return retVal
}

func (s *SoundChip) SetWaveDAC(value byte) {
	s.Wave.dacEnabled = (value>>7)&1 == 1
}

func (s *SoundChip) GetWaveDAC() byte {
	var retVal byte

	if s.Wave.dacEnabled {
		retVal |= 1 << 7
	}

	return retVal
}

func (s *SoundChip) SetWaveLengthTimer(value byte) {
	s.Wave.initLength = value
}

func (s *SoundChip) SetWaveOutput(value byte) {
	s.Wave.output = (value >> 5) & 3
}

func (s *SoundChip) GetWaveOutput() byte {
	return s.Wave.output << 5
}

func (s *SoundChip) SetWavePeriodLow(value byte) {
	s.Wave.periodLow = value
}

func (s *SoundChip) SetWavePeriodHigh(value byte) {
	if value>>7&1 == 1 {
		if s.waveLengthTimer == 0 {
			s.waveLengthTimer = LENGTH_TIMER_WAVE_MAX - uint16(s.Wave.initLength)
		}
		period := uint16(s.Wave.periodHigh)<<8 | uint16(s.Wave.periodLow)
		s.waveFreqTimer = (2048 - period) * 2
		s.waveSampleIdx = 0
		s.Global.waveEnabled = true
	}

	s.Wave.lengthEnabled = (value>>6)&1 == 1
	s.Wave.periodHigh = value & 3
}

func (s *SoundChip) GetWaveLengthEnable() byte {
	if s.Wave.lengthEnabled {
		return 1 << 6
	} else {
		return 0
	}
}

func (s *SoundChip) SetWaveRAM(addr uint16, value byte) {
	s.Wave.waveRam[addr] = value
}

func (s *SoundChip) GetWaveRAM(addr uint16) byte {
	return s.Wave.waveRam[addr]
}

func (s *SoundChip) SetNoiseLengthTimer(value byte) {
	s.Noise.initLength = value & 0x3F
}

func (s *SoundChip) SetNoiseVolumeEnvelope(value byte) {
	s.Noise.volume = value >> 4 & 0xF
	s.Noise.envDirection = value >> 3 & 1
	s.Noise.envPace = value & 7
}

func (s *SoundChip) GetNoiseVolumeEnvelope() byte {
	var retVal byte = 0

	retVal |= s.Noise.volume << 4
	retVal |= s.Noise.envDirection << 3
	retVal |= s.Noise.envPace

	return retVal
}

func (s *SoundChip) SetNoiseFreqRandomness(value byte) {
	s.Noise.clockShift = value >> 4 & 0xF
	s.Noise.lfsrWidth = value >> 3 & 1
	s.Noise.clockDivider = value & 7
}

func (s *SoundChip) GetNoiseFreqRandomness() byte {
	var retVal byte = 0

	retVal |= s.Noise.clockShift << 4
	retVal |= s.Noise.lfsrWidth << 3
	retVal |= s.Noise.clockDivider

	return retVal
}

func (s *SoundChip) SetNoiseControl(value byte) {
	if value>>7&1 == 1 {
		if s.noiseLengthTimer == 0 {
			s.noiseLengthTimer = LENGTH_TIMER_MAX - s.Noise.initLength
		}
		s.Global.noiseEnabled = true
		s.noiseLfsr = 0x7FFF
	}

	s.Noise.lengthEnabled = value>>6&1 == 1
}

func (s *SoundChip) GetNoiseControl() byte {
	if s.Noise.lengthEnabled {
		return 1 << 6
	} else {
		return 0
	}
}
