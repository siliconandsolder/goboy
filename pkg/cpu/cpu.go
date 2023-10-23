package cpu

import (
	"github.com/siliconandsolder/go-boy/pkg/bus"
	"github.com/siliconandsolder/go-boy/pkg/interrupts"
)

const (
	DMA_TRANSFER_TIME    = 160
	DMA_TRANSFER_ADDRESS = 0xFF46

	DIV_TIMER_ADDRESS = 0xFF04
	TAC_TIMER_ADDRESS = 0xFF07
)

type Cpu struct {
	// registers
	AF *Register
	BC *Register
	DE *Register
	HL *Register

	// stack pointer
	SP uint16

	// program counter
	PC uint16

	// cycles until the next command is executed
	waitCycles uint8

	// cpu has halted until next interrupts
	halt bool

	interruptEnabled bool

	dmaTransfer  bool
	dmaCountdown byte

	// CGB only
	doubleSpeed bool

	bus     *bus.Bus
	manager *interrupts.Manager
	timer   *SysTimer
}

func NewCpu(bus *bus.Bus, manager *interrupts.Manager) *Cpu {
	af := NewRegister()
	bc := NewRegister()
	de := NewRegister()
	hl := NewRegister()
	af.lower.value = 0x01
	bc.upper.value = 0x00
	bc.lower.value = 0x13
	de.upper.value = 0x00
	de.lower.value = 0xd8
	hl.upper.value = 0x01
	hl.lower.value = 0x4d

	return &Cpu{
		AF:               af,
		BC:               bc,
		DE:               de,
		HL:               hl,
		SP:               0xFFFE,
		PC:               0x0100,
		waitCycles:       0,
		halt:             false,
		interruptEnabled: false,
		doubleSpeed:      false,
		bus:              bus,
		manager:          manager,
		timer:            newSysTimer(bus),
	}
}

func (cpu *Cpu) Cycle() error {

	if err := cpu.timer.cycle(); err != nil {
		return err
	}

	if cpu.waitCycles > 0 {
		cpu.waitCycles -= 1
		return nil
	}

	if cpu.dmaTransfer {
		cpu.doDmaTransfer()
		cpu.dmaCountdown = DMA_TRANSFER_TIME
	}

	if cpu.dmaCountdown > 0 {
		cpu.dmaCountdown--
		return nil
	}

	cpu.handleInterrupts()

	// halt until an interrupt is requested
	if cpu.halt {
		cpu.waitCycles += 3
		return nil
	}

	var opCode OpCode
	var err error

	opVal := cpu.pcRead()
	if opVal == 0xCB {
		cpu.PC++
		opCode, err = GetOpCodeCB(cpu.pcRead())
		cpu.PC--
	} else {
		opCode, err = GetOpCode(opVal)
	}

	if err != nil {
		return err
	}

	opCode.execution(cpu)

	//_, err = cpu.file.WriteString(fmt.Sprintf("Opcode: %s  - PC: %d, AF: %d, BC: %d, DE: %d, HL: %d, SP: %d\n", opCode.toString, cpu.PC, cpu.AF.getAll(), cpu.BC.getAll(), cpu.DE.getAll(), cpu.HL.getAll(), cpu.SP))
	if err != nil {
		return err
	}
	//fmt.Println(fmt.Sprintf("Opcode: %s  - PC: %d, AF: %d, BC: %d, DE: %d, HL: %d, SP: %d", opCode.toString, cpu.PC, cpu.AF.getAll(), cpu.BC.getAll(), cpu.DE.getAll(), cpu.HL.getAll(), cpu.SP))
	cpu.waitCycles -= 1 // accounts for current cycle
	return nil
}

func (cpu *Cpu) setZFlag(set bool) {
	flags := cpu.AF.lower.value

	if set {
		cpu.AF.lower.value = flags | 128
	} else {
		cpu.AF.lower.value = flags & 0b01110000
	}
}

func (cpu *Cpu) setNFlag(set bool) {
	flags := cpu.AF.lower.value

	if set {
		cpu.AF.lower.value = flags | 64
	} else {
		cpu.AF.lower.value = flags & 0b10110000
	}
}

func (cpu *Cpu) setHFlag(set bool) {
	flags := cpu.AF.lower.value

	if set {
		cpu.AF.lower.value = flags | 32
	} else {
		cpu.AF.lower.value = flags & 0b11010000
	}
}

func (cpu *Cpu) setCFlag(set bool) {
	flags := cpu.AF.lower.value

	if set {
		cpu.AF.lower.value = flags | 16
	} else {
		cpu.AF.lower.value = flags & 0b11100000
	}
}

func (cpu *Cpu) getZFlag() bool {
	return cpu.AF.lower.value>>7&1 == 1
}

func (cpu *Cpu) getNFlag() bool {
	return cpu.AF.lower.value>>6&1 == 1
}

func (cpu *Cpu) getHFlag() bool {
	return cpu.AF.lower.value>>5&1 == 1
}

func (cpu *Cpu) getCFlag() bool {
	return cpu.AF.lower.value>>4&1 == 1
}

func (cpu *Cpu) pcRead() byte {
	value := cpu.bus.Read(cpu.PC)
	return value
}

func (cpu *Cpu) pcReadNext() byte {
	value := cpu.bus.Read(cpu.PC + 1)
	return value
}

func (cpu *Cpu) pcReadNext16() uint16 {
	low := cpu.bus.Read(cpu.PC + 1)
	high := cpu.bus.Read(cpu.PC + 2)
	return (uint16(high) << 8) | uint16(low)
}

func (cpu *Cpu) writeToBus(addr uint16, val byte) {
	if cpu.dmaTransfer && addr < 0xFE00 {
		return
	}

	if addr >= DIV_TIMER_ADDRESS && addr <= TAC_TIMER_ADDRESS {
		cpu.timer.write(addr, val)
		return
	}

	cpu.bus.Write(addr, val)

	if addr == 0xFF46 {
		cpu.dmaTransfer = true
	}
}

func (cpu *Cpu) readFromBus(addr uint16) byte {
	if cpu.dmaTransfer && addr < 0xFE00 {
		return 0xFF
	}

	if addr >= DIV_TIMER_ADDRESS && addr <= TAC_TIMER_ADDRESS {
		return cpu.timer.read(addr)
	}

	return cpu.bus.Read(addr)
}

func (cpu *Cpu) doDmaTransfer() {
	startAddr := uint16(cpu.bus.Read(DMA_TRANSFER_ADDRESS)) << 8
	endAddr := startAddr | 0x009F

	for i := startAddr; i <= endAddr; i++ {
		oamVal := cpu.bus.Read(i)
		cpu.bus.Write(0xFE00|(i&0x00FF), oamVal)
	}

	cpu.dmaTransfer = false
}

func (cpu *Cpu) handleInterrupts() {

	intVal := cpu.manager.GetPendingInterrupts()
	if intVal != 0 {
		cpu.halt = false
	}

	if !cpu.interruptEnabled {
		return
	}

	var pcVal uint16
	switch intVal {
	case interrupts.VBLANK:
		pcVal = 0x40
		break
	case interrupts.LCDSTAT:
		pcVal = 0x48
		break
	case interrupts.TIMER:
		pcVal = 0x50
		break
	case interrupts.SERIAL:
		pcVal = 0x58
		break
	case interrupts.JOYPAD:
		pcVal = 0x60
		break
	default:
		return
	}

	cpu.manager.DisableInterruptRequest(intVal)

	cpu.SP -= 2
	cpu.bus.Write(cpu.SP, byte(cpu.PC&0xFF)) // lower half
	cpu.bus.Write(cpu.SP+1, byte(cpu.PC>>8)) // upper half
	cpu.PC = pcVal
	cpu.interruptEnabled = false
}
