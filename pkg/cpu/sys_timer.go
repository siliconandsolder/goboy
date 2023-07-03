package cpu

import (
	"github.com/siliconandsolder/go-boy/pkg/bus"
	"github.com/siliconandsolder/go-boy/pkg/interrupts"
)

const DIV_PERIOD uint16 = 256

type SysTimer struct {
	bus       *bus.Bus
	div       byte
	divTimer  uint16
	tima      byte
	timaTimer uint16
	tma       byte
	tac       byte
	stop      bool
}

func newSysTimer(bus *bus.Bus) *SysTimer {
	return &SysTimer{
		bus:       bus,
		div:       0,
		divTimer:  0,
		tima:      0,
		timaTimer: 0,
		tma:       0,
		tac:       0,
		stop:      false,
	}
}

func (timer *SysTimer) cycle() error {
	if timer.stop {
		return nil
	}

	if timer.divTimer != DIV_PERIOD {
		timer.divTimer++
	} else {
		timer.divTimer = 0
		timer.div++
	}

	if timer.tac&4 == 4 { // TIMA enabled
		if timer.timaTimer != getTimaInterval(timer.tac&3) {
			timer.timaTimer++
			return nil
		}

		timer.tima++
		timer.timaTimer = 0

		if timer.tima == 0 {
			timer.bus.Write(bus.INTERRUPT_REQUEST, interrupts.TIMER)
			timer.tima = timer.tma
		}
	}

	return nil
}

func (timer *SysTimer) write(addr uint16, val byte) {
	switch addr {
	case 0xFF04:
		timer.div = 0
		timer.divTimer = 0
		break
	case 0xFF05:
		timer.tima = val
		break
	case 0xFF06:
		timer.tma = val
		break
	case 0xFF07:
		if timer.tac&3 != val&3 { // TIMA period changed, reset the timer
			timer.timaTimer = 0
			timer.tima = timer.tma
		}
		timer.tac = val
		break
	}
}

func (timer *SysTimer) read(addr uint16) byte {
	switch addr {
	case 0xFF04:
		return timer.div
	case 0xFF05:
		return timer.tima
	case 0xFF06:
		return timer.tma
	case 0xFF07:
		return timer.tac
	}

	return 0
}

func (timer *SysTimer) setStop(enabled bool) {
	if !enabled && timer.stop {
		timer.div = 0
	}
	timer.stop = enabled
}

func getTimaInterval(code byte) uint16 {
	switch code {
	case 0:
		return 1024
	case 1:
		return 16
	case 2:
		return 64
	case 3:
		return 256
	}

	return 0
}
