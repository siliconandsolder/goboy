package cpu

import (
	"github.com/siliconandsolder/go-boy/pkg/bus"
	"github.com/siliconandsolder/go-boy/pkg/interrupts"
)

type SysTimer struct {
	bus         *bus.Bus
	systemTimer uint16
	tima        byte
	timaTimer   uint16
	tma         byte
	tac         byte
	lastBit     byte
	cyclesToIrq byte
	stop        bool
	timaReload  bool
}

func NewSysTimer(bus *bus.Bus) *SysTimer {
	return &SysTimer{
		bus:         bus,
		systemTimer: 0,
		tima:        0,
		timaTimer:   0,
		tma:         0,
		tac:         0,
		lastBit:     0,
		cyclesToIrq: 0,
		stop:        false,
		timaReload:  false,
	}
}

func (timer *SysTimer) Cycle(cycles byte) {
	for i := byte(0); i < cycles; i++ {
		if timer.stop {
			return
		}

		timer.systemTimer++
		timer.timaReload = false

		if timer.tac&4 == 4 { // TIMA enabled
			timer.timaTimer++
			if timer.timaTimer == getTimaInterval(timer.tac&3) {
				timer.tima++
				timer.timaTimer = 0

				if timer.tima == 0 {
					timer.bus.Write(bus.INTERRUPT_REQUEST, interrupts.TIMER)
					timer.tima = timer.tma
					timer.timaReload = true
				}
			}
		}
	}
}

func (timer *SysTimer) write(addr uint16, val byte) {
	switch addr {
	case 0xFF04:
		timer.timaTimer = 0
		timer.systemTimer = 0
		break
	case 0xFF05:
		//if !timer.timaReload {
		//	timer.tima = val
		//}
		timer.tima = val
		break
	case 0xFF06:
		//if timer.timaReload {
		//	timer.tima = val
		//}
		timer.tma = val
		break
	case 0xFF07:
		if timer.tac&3 != val&3 {
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
		return byte(timer.systemTimer >> 8)
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
		timer.systemTimer = 0
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
