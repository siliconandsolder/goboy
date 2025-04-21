package cartridge

import (
	"github.com/siliconandsolder/go-boy/pkg/cartridge/rtc"
)

const (
	RAM_BANK_ADDR    = 0x5FFF
	LATCH_CLOCK_ADDR = 0x7FFF
	RTC_REG_ADDR     = 0xBFFF
)

type MBC3 struct {
	ramEnabled  bool
	romBank     byte
	lastRomBank byte
	ramBank     byte
	lastRamBank byte
	rtcActive   bool
	regToRead   byte
	prevWrite   byte

	rtcState *rtc.State
}

func NewMBC3(romInfo RomInfo, ramInfo RamInfo, rtcState *rtc.State) *MBC3 {
	return &MBC3{
		ramEnabled:  ramInfo.Size > 0,
		romBank:     0,
		lastRomBank: byte(romInfo.NumBanks - 1), // max number of banks is 128
		ramBank:     0,
		lastRamBank: ramInfo.NumBanks - 1,
		rtcActive:   false,
		regToRead:   0,
		rtcState:    rtcState,
	}
}

func (m *MBC3) Read(addr uint16) (uint32, bool) {
	if addr <= LOWER_ROM_BANK_END {
		return uint32(addr), true
	} else if addr >= UPPER_ROM_BANK_START && addr <= UPPER_ROM_BANK_END {
		return uint32(addr-0x4000) + uint32(m.romBank&m.lastRomBank)*0x4000, true
	} else if addr >= RAM_BANK_START && addr <= RAM_BANK_END {
		if !m.ramEnabled {
			return 0xFF, false
		}

		if m.rtcActive {
			m.rtcState.UpdateRTC()
			if m.rtcState.IsLatched {
				return uint32(m.rtcState.Latched.GetRegisterValue(m.regToRead)), false
			}
			return uint32(m.rtcState.Unlatched.GetRegisterValue(m.regToRead)), false
		}

		return uint32(addr-RAM_BANK_START) + 0x2000*uint32(m.ramBank&m.lastRamBank), true
	}

	return 0xFF, false
}

func (m *MBC3) Write(addr uint16, data byte) (uint32, bool) {

	if addr <= RAM_ENABLE_END {
		m.ramEnabled = data&0x0F == 0x0A
	} else if addr <= ROM_BANK_SELECT_END {
		m.romBank = data & 0x7F
		if m.romBank == 0 {
			m.romBank = 1
		}
	} else if addr <= RAM_BANK_ADDR {
		m.rtcActive = data&8 == 8
		m.ramBank = (data & 7) & m.lastRamBank
		if m.rtcActive {
			m.regToRead = data
		}
	} else if addr <= LATCH_CLOCK_ADDR {
		if data == 1 && m.prevWrite == 0 {
			m.rtcState.UpdateRTC()
			m.rtcState.IsLatched = !m.rtcState.IsLatched

			if m.rtcState.IsLatched {
				m.rtcState.Latch()
			}
		}
		m.prevWrite = data
	} else if addr <= RTC_REG_ADDR {
		if !m.ramEnabled {
			return 0, false
		}

		if m.rtcActive {
			m.rtcState.WriteToUnlatched(data)
		} else {
			return uint32(addr-RAM_BANK_START) + 0x2000*uint32(m.ramBank&m.lastRamBank), true
		}
	}

	return 0, false
}
