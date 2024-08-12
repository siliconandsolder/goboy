package audio

import "math"

const MAX_LENGTH = 64

var pulseDutyTable = [4][8]byte{
	{0, 0, 0, 0, 0, 0, 0, 1},
	{0, 0, 0, 0, 0, 0, 1, 1},
	{0, 0, 0, 0, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 0, 0},
}

type SoundChip struct {
	Global GlobalRegister
	Pulse1 PulseRegister
	Pulse2 PulseRegister
	Wave   WaveRegister
	Noise  NoiseRegister

	pulse1SweepTimer  byte
	pulse1VolumeTimer byte
	pulse1Volume      byte
	pulse1LengthTimer byte
	pulse1FreqTimer   byte
	pulse1WavePos     byte

	pulse2VolumeTimer byte
	pulse2Volume      byte
	pulse2LengthTimer byte
	pulse2FreqTimer   byte
	pulse2WavePos     byte

	frameSequencer byte
}

func (s *SoundChip) Cycle(cycles byte) {
	for i := byte(0); i < cycles; i++ {
		if s.pulse1FreqTimer == 0 {
			// reload timer with period value
			// update wave duty position
		} else {
			s.pulse1FreqTimer--
		}

		// push wave duty bit to sample slice
	}
}

func (s *SoundChip) CycleFrameSequencer() {
	if s.frameSequencer%2 == 0 {
		if s.Pulse1.PeriodHigh>>6 == 1 && s.pulse1LengthTimer > 0 {
			s.pulse1LengthTimer--
			if s.pulse1LengthTimer == 0 {
				s.Global.MasterControl &= 0xFE // disable channel 1
			}
		}

		if s.Pulse2.PeriodHigh>>6 == 1 && s.pulse2LengthTimer > 0 {
			s.pulse2LengthTimer--
			if s.pulse2LengthTimer == 0 {
				s.Global.MasterControl &= 0xFD // disable channel 2
			}
		}
	}

	if s.frameSequencer == 2 || s.frameSequencer == 6 {
		// cycle sweep
		s.pulse1SweepTimer--
		if s.pulse1SweepTimer == 0 {
			s.pulse1SweepTimer = s.Pulse1.Sweep >> 4 & 7

			if s.Pulse1.PeriodHigh>>7&1 == 1 {
				s.pulse1IterateSweep()
			}
		}
	}

	if s.frameSequencer == 7 {
		// cycle volume
		if s.Pulse1.VolumeEnv&7 != 0 {
			s.pulse1VolumeTimer--
			if s.pulse1VolumeTimer == 0 {
				s.pulse1VolumeTimer = s.Pulse1.VolumeEnv & 7

				direction := s.Pulse1.VolumeEnv >> 3 & 1
				if direction == 1 && s.pulse1Volume < 0xF {
					s.pulse1Volume++
				} else if direction == 0 && s.pulse1Volume > 0 {
					s.pulse1Volume--
				}
			}
		}

		if s.Pulse2.VolumeEnv&7 != 0 {
			s.pulse2VolumeTimer--
			if s.pulse2VolumeTimer == 0 {
				s.pulse2VolumeTimer = s.Pulse2.VolumeEnv & 7

				direction := s.Pulse2.VolumeEnv >> 3 & 1
				if direction == 1 && s.pulse2Volume < 0xF {
					s.pulse2Volume++
				} else if direction == 0 && s.pulse2Volume > 0 {
					s.pulse2Volume--
				}
			}
		}
	}

	s.frameSequencer++
	s.frameSequencer &= 7
}

func (s *SoundChip) pulse1IterateSweep() {
	curPeriod := uint16(s.Pulse1.PeriodHigh&7)<<8 | uint16(s.Pulse1.PeriodLow)
	stepVal := curPeriod / uint16(math.Pow(2.0, float64(s.Pulse1.Sweep&7)))

	var newPeriod uint16
	direction := s.Pulse1.Sweep >> 3 & 1
	if direction != 0 {
		newPeriod -= stepVal
	} else {
		newPeriod += stepVal
	}

	s.Pulse1.PeriodLow = byte(newPeriod & 0xFF)
	s.Pulse1.PeriodHigh = (s.Pulse1.PeriodHigh & 0xF8) | (byte(newPeriod>>8) & 7)

	if newPeriod > 0x7FF {
		s.Global.MasterControl &= 0xFE // turn off pulse 1 channel
	}
}
