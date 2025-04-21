package rtc

import "time"

const CYCLES_PER_SECOND = 4194304

type State struct {
	Unlatched *registers
	Latched   *registers
	IsLatched bool
	cycles    uint64
	lastTime  time.Time
}

type StateSnapshot struct {
	Seconds         byte
	Minutes         byte
	Hours           byte
	DaysLow         byte
	DaysHigh        byte
	LatchedSeconds  byte
	LatchedMinutes  byte
	LatchedHours    byte
	LatchedDaysLow  byte
	LatchedDaysHigh byte
	Timestamp       int64
}

func NewState() *State {
	return &State{
		Unlatched: &registers{
			S:  0,
			M:  0,
			H:  0,
			DL: 0,
			DH: 0,
		},
		Latched: &registers{
			S:  0,
			M:  0,
			H:  0,
			DL: 0,
			DH: 0,
		},
		IsLatched: false,
		cycles:    0,
		lastTime:  time.Now(),
	}
}

func (s *State) AddCycles(cycles byte) {
	s.cycles += uint64(cycles)
}

func (s *State) UpdateRTC() {
	if s.Unlatched.DH&0x40 == 0x40 { // halt flag
		s.cycles = 0
		return
	}

	for s.cycles >= CYCLES_PER_SECOND {
		s.cycles -= CYCLES_PER_SECOND
		now := time.Now()

		s.Unlatched.S += byte((now.Second() - s.lastTime.Second()) & 59)
		if now.Second() < s.lastTime.Second() {
			s.Unlatched.M++
		}

		s.Unlatched.M += byte((now.Minute() - s.lastTime.Minute()) & 59)
		if now.Minute() < s.lastTime.Minute() {
			s.Unlatched.H++
		}

		s.Unlatched.H += byte((now.Hour() - s.lastTime.Hour()) & 23)
		if now.Hour() < s.lastTime.Hour() {
			s.Unlatched.DL++
		}

		if (s.Unlatched.DL == 0 || s.Unlatched.DL > byte(now.YearDay()&255)) && s.Unlatched.DH&1 == 1 { // day overflow
			s.Unlatched.DH |= 0x80
		}

		s.Unlatched.DL += byte((now.YearDay() - s.lastTime.YearDay()) & 255)
		s.Unlatched.DH = (s.Unlatched.DH & 0xC0) | byte(now.YearDay()>>7&1)

		s.lastTime = now
	}
}

func (s *State) WriteToUnlatched(val byte) {
	switch val {
	case 0x08:
		s.Unlatched.S = val & 0x3F
		s.cycles = 0
	case 0x09:
		s.Unlatched.M = val & 0x3F
	case 0x0A:
		s.Unlatched.H = val & 0x1F
	case 0x0B:
		s.Unlatched.DL = val
	case 0x0C:
		if val>>6&1 == 1 && s.Unlatched.DH>>6&1 == 1 { // RTC enabled
			s.lastTime = time.Now()
		}
		s.Unlatched.DH = val & 0xC1
	default:
		panic("unrecognized value: " + string(val))
	}
}

func (s *State) Latch() {
	s.Latched.S = s.Unlatched.S
	s.Latched.M = s.Unlatched.M
	s.Latched.H = s.Unlatched.H
	s.Latched.DL = s.Unlatched.DL
	s.Latched.DH = s.Unlatched.DH
	s.IsLatched = true
}

func (s *State) GetSnapshot() *StateSnapshot {
	return &StateSnapshot{
		Seconds:         s.Unlatched.S,
		Minutes:         s.Unlatched.M,
		Hours:           s.Unlatched.H,
		DaysLow:         s.Unlatched.DL,
		DaysHigh:        s.Unlatched.DH,
		LatchedSeconds:  s.Latched.S,
		LatchedMinutes:  s.Latched.M,
		LatchedHours:    s.Latched.H,
		LatchedDaysLow:  s.Latched.DL,
		LatchedDaysHigh: s.Latched.DH,
		Timestamp:       time.Now().Unix(),
	}
}

func (s *State) FromSnapshot(snapshot *StateSnapshot) {
	s.Unlatched.S = snapshot.Seconds
	s.Unlatched.M = snapshot.Minutes
	s.Unlatched.H = snapshot.Hours
	s.Unlatched.DL = snapshot.DaysLow
	s.Unlatched.DH = snapshot.DaysHigh
	s.Latched.S = snapshot.LatchedSeconds
	s.Latched.M = snapshot.LatchedMinutes
	s.Latched.H = snapshot.LatchedHours
	s.Latched.DL = snapshot.LatchedDaysLow
	s.Latched.DH = snapshot.LatchedDaysHigh
	s.lastTime = time.Unix(snapshot.Timestamp, 0)
}
