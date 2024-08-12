package audio

type GlobalRegister struct {
	MasterVolume  byte // NR50
	Panning       byte // NR51
	MasterControl byte // NR52
}

type PulseRegister struct {
	Sweep      byte // NR10 - Channel 1 only
	DutyLength byte // NR11/NR21
	VolumeEnv  byte // NR12/NR22
	PeriodLow  byte // NR13/NR23
	PeriodHigh byte // NR14/NR24
}

type WaveRegister struct {
	DacEnabled byte   // NR30
	Length     byte   // NR31
	Output     byte   // NR32
	PeriodLow  byte   // NR33
	PeriodHigh byte   // NR34
	WaveRam    []byte // NR35-NR40
}

type NoiseRegister struct {
	Length    byte // NR41
	Volume    byte // NR42
	Frequency byte // NR43
	Control   byte // NR44
}
