package audio

import "github.com/veandco/go-sdl2/sdl"

const LENGTH_TIMER_MAX = 64
const LENGTH_TIMER_WAVE_MAX = 256
const CYCLES_PER_SAMPLE = 87
const CYCLE_SEQUENCER_MAX = 8192

type SoundChip struct {
	Global GlobalRegister
	Pulse1 pulseRegister
	Pulse2 pulseRegister
	Wave   waveRegister
	Noise  noiseRegister

	frameSequencer    byte
	cyclesToSequencer uint16
	cyclesToSample    byte
	player            *Player
}

func NewSoundChip(p *Player) *SoundChip {
	return &SoundChip{
		Global: GlobalRegister{},
		Pulse1: pulseRegister{
			isChannel1: true,
		},
		Pulse2: pulseRegister{},
		Wave: waveRegister{
			ram: make([]byte, 0x20),
		},
		Noise:             noiseRegister{},
		frameSequencer:    0,
		cyclesToSequencer: 0,
		cyclesToSample:    CYCLES_PER_SAMPLE,
		player:            p,
	}
}

func (s *SoundChip) Cycle(cycles byte) {
	for i := byte(0); i < cycles; i++ {

		s.cyclesToSequencer++
		if s.cyclesToSequencer == CYCLE_SEQUENCER_MAX {
			s.CycleFrameSequencer()
			s.cyclesToSequencer = 0
		}
		s.Pulse1.cycleFrequencyTimer()
		s.Pulse2.cycleFrequencyTimer()
		s.Wave.cycleFrequencyTimer()
		s.Noise.cycleFrequencyTimer()

		s.cyclesToSample--
		if s.cyclesToSample == 0 {
			s.cyclesToSample = CYCLES_PER_SAMPLE

			var pulse1SampleL byte = 0
			var pulse2SampleL byte = 0
			var waveSampleL byte = 0
			var noiseSampleL byte = 0

			var pulse1SampleR byte = 0
			var pulse2SampleR byte = 0
			var waveSampleR byte = 0
			var noiseSampleR byte = 0

			if s.Global.audioEnabled {
				if s.Pulse1.enabled && s.Pulse1.dacEnabled {
					if s.Global.pulse1Left {
						pulse1SampleL = s.Pulse1.getSample()
					}
					if s.Global.pulse1Right {
						pulse1SampleR = s.Pulse1.getSample()
					}
				}

				if s.Pulse2.enabled && s.Pulse2.dacEnabled {
					if s.Global.pulse2Left {
						pulse2SampleL = s.Pulse2.getSample()
					}
					if s.Global.pulse2Right {
						pulse2SampleR = s.Pulse2.getSample()
					}
				}

				if s.Wave.enabled && s.Wave.dacEnabled {
					if s.Global.waveLeft {
						waveSampleL = s.Wave.getSample()
					}
					if s.Global.waveRight {
						waveSampleR = s.Wave.getSample()
					}
				}

				if s.Noise.enabled {
					if s.Global.noiseLeft {
						noiseSampleL = s.Noise.getSample()
					}
					if s.Global.noiseRight {
						noiseSampleR = s.Noise.getSample()
					}
				}
			}

			mixedSampleLeft := pulse1SampleL + pulse2SampleL + waveSampleL + noiseSampleL
			mixedSampleRight := pulse1SampleR + pulse2SampleR + waveSampleR + noiseSampleR
			s.player.SendSample(stereoSample{
				leftSample:  mixedSampleLeft,
				rightSample: mixedSampleRight,
			})
			for len(s.player.channel) > AUDIO_FREQUENCY/30 { // two buffers' worth
				sdl.Delay(1)
			}
		}
	}
}

func (s *SoundChip) CycleFrameSequencer() {
	if s.frameSequencer%2 == 0 {
		s.Pulse1.cycleLengthTimer()
		s.Pulse2.cycleLengthTimer()
		s.Wave.cycleLengthTimer()
		s.Noise.cycleLengthTimer()
	}

	if s.frameSequencer == 2 || s.frameSequencer == 6 {
		s.Pulse1.cycleSweepTimer()
	}

	if s.frameSequencer == 7 {
		// cycle volume
		s.Pulse1.cycleVolumeTimer()
		s.Pulse2.cycleVolumeTimer()
		s.Noise.cycleVolumeTimer()
	}

	s.frameSequencer++
	s.frameSequencer &= 7
}

func (s *SoundChip) IsOn() bool {
	return s.Global.audioEnabled
}

func (s *SoundChip) SetMasterControl(value byte) {
	audioEnabled := value>>7&1 == 1
	if audioEnabled && !s.Global.audioEnabled {
		s.frameSequencer = 0
	}

	s.Global.audioEnabled = audioEnabled

	if !s.Global.audioEnabled {
		s.Pulse1.enabled = false
		s.Pulse2.enabled = false
		s.Wave.enabled = false
		s.Noise.enabled = false

		s.Pulse1.clear()
		s.Pulse2.clear()
		s.Wave.clear()
		s.Noise.clear()
		s.Global.clear()
	}
}

func (s *SoundChip) GetMasterControl() byte {
	var retVal byte = 0
	if s.Global.audioEnabled {
		retVal |= 1 << 7
	}

	if s.Noise.enabled {
		retVal |= 1 << 3
	}

	if s.Wave.enabled {
		retVal |= 1 << 2
	}

	if s.Pulse2.enabled {
		retVal |= 1 << 1
	}

	if s.Pulse1.enabled {
		retVal |= 1
	}

	return retVal | registerMasks[nr52]
}

func (s *SoundChip) SetMasterVolume(value byte) {
	s.Global.vinLeft = value >> 7 & 1
	s.Global.leftVolume = value >> 4 & 7
	s.Global.vinRight = value >> 3 & 1
	s.Global.rightVolume = value & 7
}

func (s *SoundChip) GetMasterVolume() byte {
	var retVal byte = 0

	retVal |= s.Global.vinLeft << 7
	retVal |= s.Global.leftVolume << 4
	retVal |= s.Global.vinRight << 3
	retVal |= s.Global.rightVolume

	return retVal | registerMasks[nr50]
}

func (s *SoundChip) SetMasterPanning(value byte) {
	s.Global.noiseLeft = value>>7&1 == 1
	s.Global.waveLeft = value>>6&1 == 1
	s.Global.pulse2Left = value>>5&1 == 1
	s.Global.pulse1Left = value>>4&1 == 1

	s.Global.noiseRight = value>>3&1 == 1
	s.Global.waveRight = value>>2&1 == 1
	s.Global.pulse2Right = value>>1&1 == 1
	s.Global.pulse1Right = value&1 == 1
}

func (s *SoundChip) GetMasterPanning() byte {
	var retVal byte = 0

	if s.Global.noiseLeft {
		retVal |= 1 << 7
	}

	if s.Global.waveLeft {
		retVal |= 1 << 6
	}

	if s.Global.pulse2Left {
		retVal |= 1 << 5
	}

	if s.Global.pulse1Left {
		retVal |= 1 << 4
	}

	if s.Global.noiseRight {
		retVal |= 1 << 3
	}

	if s.Global.waveRight {
		retVal |= 1 << 2
	}

	if s.Global.pulse2Right {
		retVal |= 1 << 1
	}

	if s.Global.pulse1Right {
		retVal |= 1
	}

	return retVal | registerMasks[nr51]
}

func (s *SoundChip) SetPulse1Sweep(value byte) {
	s.Pulse1.setSweep(value)
}

func (s *SoundChip) GetPulse1Sweep() byte {
	return s.Pulse1.getSweep() | registerMasks[nr10]
}

func (s *SoundChip) SetPulse1LengthDuty(value byte) {
	s.Pulse1.setLengthDuty(value)
}

func (s *SoundChip) GetPulse1LengthDuty() byte {
	return s.Pulse1.getLengthDuty() | registerMasks[nr11]
}

func (s *SoundChip) SetPulse1VolumeEnvelope(value byte) {
	s.Pulse1.setVolumeEnvelope(value)
}

func (s *SoundChip) GetPulse1VolumeEnvelope() byte {
	return s.Pulse1.getVolumeEnvelope() | registerMasks[nr12]
}

func (s *SoundChip) SetPulse1PeriodLow(value byte) {
	s.Pulse1.setPeriodLow(value)
}

func (s *SoundChip) SetPulse1PeriodHigh(value byte) {
	s.Pulse1.setPeriodHigh(value, s.frameSequencer%2 == 0)
}

func (s *SoundChip) GetPulse1PeriodHigh() byte {
	return s.Pulse1.getPeriodHigh() | registerMasks[nr14]
}

func (s *SoundChip) SetPulse2LengthDuty(value byte) {
	s.Pulse2.setLengthDuty(value)
}

func (s *SoundChip) GetPulse2LengthDuty() byte {
	return s.Pulse2.getLengthDuty() | registerMasks[nr21]
}

func (s *SoundChip) SetPulse2VolumeEnvelope(value byte) {
	s.Pulse2.setVolumeEnvelope(value)
}

func (s *SoundChip) GetPulse2VolumeEnvelope() byte {
	return s.Pulse2.getVolumeEnvelope() | registerMasks[nr22]
}

func (s *SoundChip) SetPulse2PeriodLow(value byte) {
	s.Pulse2.setPeriodLow(value)
}

func (s *SoundChip) SetPulse2PeriodHigh(value byte) {
	s.Pulse2.setPeriodHigh(value, s.frameSequencer%2 == 0)
}

func (s *SoundChip) GetPulse2PeriodHigh() byte {
	return s.Pulse2.getPeriodHigh() | registerMasks[nr24]
}

func (s *SoundChip) SetWaveDAC(value byte) {
	s.Wave.dacEnabled = (value>>7)&1 == 1
	if !s.Wave.dacEnabled {
		s.Wave.enabled = false
	}
}

func (s *SoundChip) GetWaveDAC() byte {
	return s.Wave.getDAC() | registerMasks[nr30]
}

func (s *SoundChip) SetWaveLengthTimer(value byte) {
	s.Wave.initLength = value
	s.Wave.lengthTimer = LENGTH_TIMER_WAVE_MAX - uint16(value)
}

func (s *SoundChip) SetWaveOutput(value byte) {
	s.Wave.output = (value >> 5) & 3
}

func (s *SoundChip) GetWaveOutput() byte {
	return (s.Wave.output << 5) | registerMasks[nr32]
}

func (s *SoundChip) SetWavePeriodLow(value byte) {
	s.Wave.periodLow = value
}

func (s *SoundChip) SetWavePeriodHigh(value byte) {
	s.Wave.setPeriodHigh(value)
}

func (s *SoundChip) GetWaveLengthEnable() byte {
	return s.Wave.getLengthEnabled() | registerMasks[nr34]
}

func (s *SoundChip) SetWaveRAM(addr uint16, value byte) {
	s.Wave.ram[addr] = value
}

func (s *SoundChip) GetWaveRAM(addr uint16) byte {
	return s.Wave.ram[addr]
}

func (s *SoundChip) SetNoiseLengthTimer(value byte) {
	s.Noise.initLength = value & 0x3F
	s.Noise.lengthTimer = LENGTH_TIMER_MAX - s.Noise.initLength
}

func (s *SoundChip) SetNoiseVolumeEnvelope(value byte) {
	s.Noise.setVolumeEnvelope(value)
}

func (s *SoundChip) GetNoiseVolumeEnvelope() byte {
	return s.Noise.getVolumeEnvelope() | registerMasks[nr42]
}

func (s *SoundChip) SetNoiseFreqRandomness(value byte) {
	s.Noise.setFreqRandomness(value)
}

func (s *SoundChip) GetNoiseFreqRandomness() byte {
	return s.Noise.getFreqRandomness() | registerMasks[nr43]
}

func (s *SoundChip) SetNoiseControl(value byte) {
	s.Noise.setNoiseControl(value)
}

func (s *SoundChip) GetNoiseControl() byte {
	return s.Noise.getNoiseControl() | registerMasks[nr44]
}
