package interrupts

const (
	VBLANK  byte = 1
	LCDSTAT byte = 2
	TIMER   byte = 4
	SERIAL  byte = 8
	JOYPAD  byte = 16
)

type interrupt struct {
	request bool
	enabled bool
}

/*
Bit 0: VBlank   Interrupt Enable  (INT $40)  (1=Enable)
Bit 1: LCD STAT Interrupt Enable  (INT $48)  (1=Enable)
Bit 2: Timer    Interrupt Enable  (INT $50)  (1=Enable)
Bit 3: Serial   Interrupt Enable  (INT $58)  (1=Enable)
Bit 4: Joypad   Interrupt Enable  (INT $60)  (1=Enable)
*/

type Manager struct {
	vBlank  interrupt
	lcdStat interrupt
	timer   interrupt
	serial  interrupt
	joypad  interrupt
}

func NewManager() *Manager {
	return &Manager{
		vBlank:  interrupt{},
		lcdStat: interrupt{},
		timer:   interrupt{},
		serial:  interrupt{},
		joypad:  interrupt{},
	}
}

func (m *Manager) GetPendingInterrupts() byte {
	if m.vBlank.request && m.vBlank.enabled {
		return VBLANK
	} else if m.lcdStat.request && m.lcdStat.enabled {
		return LCDSTAT
	} else if m.timer.request && m.timer.enabled {
		return TIMER
	} else if m.serial.request && m.serial.enabled {
		return SERIAL
	} else if m.joypad.request && m.joypad.enabled {
		return JOYPAD
	}

	return 0
}

func (m *Manager) DisableInterruptRequest(val byte) {
	switch val {
	case VBLANK:
		m.vBlank.request = false
		break
	case LCDSTAT:
		m.lcdStat.request = false
		break
	case TIMER:
		m.timer.request = false
		break
	case SERIAL:
		m.serial.request = false
		break
	case JOYPAD:
		m.joypad.request = false
		break
	default:
		break
	}
}

func (m *Manager) EnableAllInterrupts() {
	m.vBlank.enabled = true
	m.lcdStat.enabled = true
	m.timer.enabled = true
	m.serial.enabled = true
	m.joypad.enabled = true
}

func (m *Manager) SetInterruptEnable(val byte) {
	m.vBlank.enabled = val&VBLANK == VBLANK
	m.lcdStat.enabled = val&LCDSTAT == LCDSTAT
	m.timer.enabled = val&TIMER == TIMER
	m.serial.enabled = val&SERIAL == SERIAL
	m.joypad.enabled = val&JOYPAD == JOYPAD
}

func (m *Manager) ToggleInterruptRequest(val byte) {
	if val&VBLANK == VBLANK {
		m.vBlank.request = true
	}
	if val&LCDSTAT == LCDSTAT {
		m.lcdStat.request = true
	}
	if val&TIMER == TIMER {
		m.timer.request = true
	}
	if val&SERIAL == SERIAL {
		m.serial.request = true
	}
	if val&JOYPAD == JOYPAD {
		m.joypad.request = true
	}
}

func (m *Manager) SetInterruptRequest(val byte) {
	m.vBlank.request = val&VBLANK == VBLANK
	m.lcdStat.request = val&LCDSTAT == LCDSTAT
	m.timer.request = val&TIMER == TIMER
	m.serial.request = val&SERIAL == SERIAL
	m.joypad.request = val&JOYPAD == JOYPAD
}

func (m *Manager) GetInterruptRequests() byte {
	var intVal byte = 0
	if m.vBlank.request {
		intVal |= VBLANK
	}
	if m.lcdStat.request {
		intVal |= LCDSTAT
	}
	if m.timer.request {
		intVal |= TIMER
	}
	if m.serial.request {
		intVal |= SERIAL
	}
	if m.joypad.request {
		intVal |= JOYPAD
	}

	return intVal
}

func (m *Manager) GetEnabledInterrupts() byte {
	var intVal byte = 0
	if m.vBlank.enabled {
		intVal |= VBLANK
	}
	if m.lcdStat.enabled {
		intVal |= LCDSTAT
	}
	if m.timer.enabled {
		intVal |= TIMER
	}
	if m.serial.enabled {
		intVal |= SERIAL
	}
	if m.joypad.enabled {
		intVal |= JOYPAD
	}

	return intVal
}
