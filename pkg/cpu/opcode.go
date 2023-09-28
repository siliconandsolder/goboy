package cpu

import "fmt"

type Execution func(*Cpu)

type OpCode struct {
	execution Execution
	toString  string
}

func GetOpCode(code byte) (OpCode, error) {
	switch code {
	case 0x00:
		return OpCode{
			execution: func(cpu *Cpu) {
				cpu.PC += 1
				cpu.waitCycles += 4
			},
			toString: "NOP",
		}, nil
	case 0x01:
		return OpCode{
			execution: func(cpu *Cpu) {
				cpu.BC.setAll(cpu.pcReadNext16())
				cpu.PC += 3
				cpu.waitCycles += 12
			},
			toString: "LD BC,u16",
		}, nil
	case 0x02:
		return OpCode{
			execution: func(cpu *Cpu) {
				cpu.writeToBus(cpu.BC.getAll(), cpu.AF.upper.value)
				cpu.PC += 1
				cpu.waitCycles += 8
			},
			toString: "LD (BC),A",
		}, nil
	case 0x03:
		return OpCode{
			execution: func(cpu *Cpu) {
				cpu.BC.setAll(cpu.BC.getAll() + 1)
				cpu.PC += 1
				cpu.waitCycles += 8
			},
			toString: "INC BC",
		}, nil
	case 0x04:
		return OpCode{
			execution: func(c *Cpu) {
				inc8(c, c.BC.upper)
			},
			toString: "INC B",
		}, nil
	case 0x05:
		return OpCode{
			execution: func(c *Cpu) {
				dec8(c, c.BC.upper)
			},
			toString: "DEC B",
		}, nil
	case 0x06:
		return OpCode{
			execution: func(c *Cpu) {
				c.BC.upper.value = c.pcReadNext()
				c.PC += 2
				c.waitCycles += 8
			},
			toString: "LD B,u8",
		}, nil
	case 0x07:
		return OpCode{
			execution: func(c *Cpu) {
				// rotate A
				orgVal := c.AF.upper.value
				c.AF.upper.value <<= 1

				if orgVal&0x80 == 0x80 {
					c.AF.upper.value |= 1
				}

				c.setCFlag(orgVal&0x80 == 0x80)
				c.setNFlag(false)
				c.setZFlag(false)
				c.setHFlag(false)

				c.PC += 1
				c.waitCycles += 4
			},
			toString: "RLCA",
		}, nil
	case 0x08:
		return OpCode{
			execution: func(c *Cpu) {
				addr := c.pcReadNext16()
				c.writeToBus(addr, byte(c.SP&0x00FF))
				c.writeToBus(addr+1, byte(c.SP>>8))

				c.PC += 3
				c.waitCycles += 20
			},
			toString: "LD (u16),SP",
		}, nil
	case 0x09:
		return OpCode{
			execution: func(c *Cpu) {
				addHL(c, c.BC.getAll())
			},
			toString: "ADD HL,BC",
		}, nil
	case 0x0A:
		return OpCode{
			execution: func(c *Cpu) {
				load(c.AF.upper, c.readFromBus(c.BC.getAll()))

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD A,(BC)",
		}, nil
	case 0x0B:
		return OpCode{
			execution: func(c *Cpu) {
				c.BC.setAll(c.BC.getAll() - 1)

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "DEC BC",
		}, nil
	case 0x0C: // INC C
		return OpCode{
			execution: func(c *Cpu) {
				inc8(c, c.BC.lower)
			},
			toString: "INC C",
		}, nil
	case 0x0D: // DEC C
		return OpCode{
			execution: func(c *Cpu) {
				dec8(c, c.BC.lower)
			},
			toString: "DEC C",
		}, nil
	case 0x0E: // LD C, u8
		return OpCode{
			execution: func(c *Cpu) {
				c.BC.lower.value = c.pcReadNext()

				c.PC += 2
				c.waitCycles += 8
			},
			toString: "LD C, u8",
		}, nil
	case 0x0F: // RRCA
		return OpCode{
			execution: func(c *Cpu) {
				val := (c.AF.upper.value >> 1) | (c.AF.upper.value << 7)
				c.setZFlag(false)
				c.setNFlag(false)
				c.setHFlag(false)
				c.setCFlag(c.AF.upper.value&0x01 == 0x01)
				c.AF.upper.value = val

				c.PC += 1
				c.waitCycles += 4
			},
			toString: "RRCA",
		}, nil
	case 0x10: // STOP 0
		return OpCode{
			execution: func(c *Cpu) {
				// signal STOP to timer
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "STOP 0",
		}, nil
	case 0x11: // LD DE, u16
		return OpCode{
			execution: func(c *Cpu) {
				c.DE.setAll(c.pcReadNext16())
				c.PC += 3
				c.waitCycles += 12
			},
			toString: "LD DE, u16",
		}, nil
	case 0x12: // LD (DE), A
		return OpCode{
			execution: func(c *Cpu) {
				c.writeToBus(c.DE.getAll(), c.AF.upper.value)
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD (DE), A",
		}, nil
	case 0x13: // INC DE
		return OpCode{
			execution: func(c *Cpu) {
				c.DE.setAll(c.DE.getAll() + 1)
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "INC DE",
		}, nil
	case 0x14: // INC D
		return OpCode{
			execution: func(c *Cpu) {
				inc8(c, c.DE.upper)
			},
			toString: "INC D",
		}, nil
	case 0x15: // DEC D
		return OpCode{
			execution: func(c *Cpu) {
				dec8(c, c.DE.upper)
			},
			toString: "DEC D",
		}, nil
	case 0x16: // LD D, u8
		return OpCode{
			execution: func(c *Cpu) {
				c.DE.upper.value = c.pcReadNext()
				c.PC += 2
				c.waitCycles += 8
			},
			toString: "LD D, u8",
		}, nil
	case 0x17: // RLA
		return OpCode{
			execution: func(c *Cpu) {
				carryVal := 0
				if c.getCFlag() {
					carryVal = 1
				}
				val := c.AF.upper.value<<1 | uint8(carryVal)

				c.setNFlag(false)
				c.setZFlag(false)
				c.setHFlag(false)
				c.setCFlag(c.AF.upper.value>>7 == 1)
				c.AF.upper.value = val

				c.PC += 1
				c.waitCycles += 4
			},
			toString: "RLA",
		}, nil
	case 0x18: // JR r8
		return OpCode{
			execution: func(c *Cpu) {
				jumpRelative(c, true)
			},
			toString: "JR r8",
		}, nil
	case 0x19: // ADD HL,DE
		return OpCode{
			execution: func(c *Cpu) {
				addHL(c, c.DE.getAll())
			},
			toString: "ADD HL,DE",
		}, nil
	case 0x1A: // LD A,(DE)
		return OpCode{
			execution: func(c *Cpu) {
				load(c.AF.upper, c.readFromBus(c.DE.getAll()))
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD A,(DE)",
		}, nil
	case 0x1B: // DEC DE
		return OpCode{
			execution: func(c *Cpu) {
				c.DE.setAll(c.DE.getAll() - 1)
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "DEC DE",
		}, nil
	case 0x1C: // INC E
		return OpCode{
			execution: func(c *Cpu) {
				inc8(c, c.DE.lower)
			},
			toString: "INC E",
		}, nil
	case 0x1D: // DEC E
		return OpCode{
			execution: func(c *Cpu) {
				dec8(c, c.DE.lower)
			},
			toString: "DEC E",
		}, nil
	case 0x1E: // LD E,d8
		return OpCode{
			execution: func(c *Cpu) {
				c.DE.lower.value = c.pcReadNext()
				c.PC += 2
				c.waitCycles += 8
			},
			toString: "LD E,d8",
		}, nil
	case 0x1F: // RRA
		return OpCode{
			execution: func(c *Cpu) {
				newVal := c.AF.upper.value >> 1
				if c.getCFlag() {
					newVal |= 0x80
				}
				c.setNFlag(false)
				c.setZFlag(false)
				c.setHFlag(false)
				c.setCFlag(c.AF.upper.value&1 == 1)
				c.AF.upper.value = newVal

				c.PC += 1
				c.waitCycles += 4
			},
			toString: "RRA",
		}, nil
	case 0x20: // JR NZ,r8
		return OpCode{
			execution: func(c *Cpu) {
				jumpRelative(c, !c.getZFlag())
			},
			toString: "JR NZ,r8",
		}, nil
	case 0x21: // LD HL,u16
		return OpCode{
			execution: func(c *Cpu) {
				c.HL.setAll(c.pcReadNext16())
				c.PC += 3
				c.waitCycles += 12
			},
			toString: "LD HL,u16",
		}, nil
	case 0x22: // LD (HL+),A
		return OpCode{
			execution: func(c *Cpu) {
				c.writeToBus(c.HL.getAll(), c.AF.upper.value)
				c.HL.setAll(c.HL.getAll() + 1)
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD (HL+),A",
		}, nil
	case 0x23: // INC HL
		return OpCode{
			execution: func(c *Cpu) {
				c.HL.setAll(c.HL.getAll() + 1)
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "INC HL",
		}, nil
	case 0x24: // INC H
		return OpCode{
			execution: func(c *Cpu) {
				inc8(c, c.HL.upper)
			},
			toString: "INC H",
		}, nil
	case 0x25: // DEC H
		return OpCode{
			execution: func(c *Cpu) {
				dec8(c, c.HL.upper)
			},
			toString: "DEC H",
		}, nil
	case 0x26: // LD H,d8
		return OpCode{
			execution: func(c *Cpu) {
				c.HL.upper.value = c.pcReadNext()
				c.PC += 2
				c.waitCycles += 8
			},
			toString: "LD H,d8",
		}, nil
	case 0x27: // DAA
		return OpCode{
			execution: func(c *Cpu) {
				// TODO: fix DAA
				corr := int16(c.AF.upper.value)
				if c.getNFlag() {
					if c.getHFlag() {
						corr = (corr - 6) & 0xFF
					}
					if c.getCFlag() {
						corr -= 0x60
					}
				} else {
					if c.getHFlag() || ((corr & 0x0F) > 0x09) {
						corr += 0x06
					}
					if c.getCFlag() || (corr > 0x9F) {
						corr += 0x60
						//c.setCFlag(true)
					}
				}

				c.setHFlag(false)

				if corr&0x100 == 0x100 {
					c.setCFlag(true)
				}

				c.AF.upper.value = byte(corr & 0xFF)
				c.setZFlag(c.AF.upper.value == 0)

				c.PC += 1
				c.waitCycles += 4
			},
			toString: "DAA",
		}, nil
	case 0x28: // JR Z,r8
		return OpCode{
			execution: func(c *Cpu) {
				jumpRelative(c, c.getZFlag())
			},
			toString: "JR Z,r8",
		}, nil
	case 0x29: // ADD HL,HL
		return OpCode{
			execution: func(c *Cpu) {
				addHL(c, c.HL.getAll())
			},
			toString: "ADD HL,HL",
		}, nil
	case 0x2A: // LD A,(HL+)
		return OpCode{
			execution: func(c *Cpu) {
				c.AF.upper.value = c.readFromBus(c.HL.getAll())
				c.HL.setAll(c.HL.getAll() + 1)
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD A,(HL+)",
		}, nil
	case 0x2B: // DEC HL
		return OpCode{
			execution: func(c *Cpu) {
				c.HL.setAll(c.HL.getAll() - 1)
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "DEC HL",
		}, nil
	case 0x2C: // INC L
		return OpCode{
			execution: func(c *Cpu) {
				inc8(c, c.HL.lower)
			},
			toString: "INC L",
		}, nil
	case 0x2D: // DEC L
		return OpCode{
			execution: func(c *Cpu) {
				dec8(c, c.HL.lower)
			},
			toString: "DEC L",
		}, nil
	case 0x2E: // LD L,d8
		return OpCode{
			execution: func(c *Cpu) {
				c.HL.lower.value = c.pcReadNext()
				c.PC += 2
				c.waitCycles += 8
			},
			toString: "LD L,d8",
		}, nil
	case 0x2F: // CPL
		return OpCode{
			execution: func(c *Cpu) {
				c.AF.upper.value = c.AF.upper.value ^ 0xFF
				c.setNFlag(true)
				c.setHFlag(true)
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "CPL",
		}, nil
	case 0x30: // JR NC,r8
		return OpCode{
			execution: func(c *Cpu) {
				jumpRelative(c, !c.getCFlag())
			},
			toString: "JR NC,r8",
		}, nil
	case 0x31: // LD SP,u16
		return OpCode{
			execution: func(c *Cpu) {
				c.SP = c.pcReadNext16()

				c.PC += 3
				c.waitCycles += 12
			},
			toString: "LD SP,u16",
		}, nil
	case 0x32: // LD (HL-),A
		return OpCode{
			execution: func(c *Cpu) {
				c.writeToBus(c.HL.getAll(), c.AF.upper.value)
				c.HL.setAll(c.HL.getAll() - 1)
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD (HL-),A",
		}, nil
	case 0x33: // INC SP
		return OpCode{
			execution: func(c *Cpu) {
				c.SP += 1

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "INC SP",
		}, nil
	case 0x34: // INC (HL)
		return OpCode{
			execution: func(c *Cpu) {
				addr := c.HL.getAll()
				val := c.readFromBus(addr) + 1
				c.writeToBus(addr, val)
				c.setNFlag(false)
				c.setZFlag(val == 0)
				c.setHFlag(val&0xF == 0)

				c.PC += 1
				c.waitCycles += 12
			},
			toString: "INC (HL)",
		}, nil
	case 0x35: // DEC (HL)
		return OpCode{
			execution: func(c *Cpu) {
				addr := c.HL.getAll()
				val := c.readFromBus(addr) - 1

				c.setZFlag(val == 0)
				c.setNFlag(true)
				c.setHFlag((val & 0x0F) == 0x0F)

				c.writeToBus(addr, val)

				c.PC += 1
				c.waitCycles += 12
			},
			toString: "DEC (HL)",
		}, nil
	case 0x36: // LD (HL),d8
		return OpCode{
			execution: func(c *Cpu) {
				c.writeToBus(c.HL.getAll(), c.pcReadNext())
				c.PC += 2
				c.waitCycles += 12
			},
			toString: "LD (HL),d8",
		}, nil
	case 0x37: // SCF
		return OpCode{
			execution: func(c *Cpu) {
				c.setNFlag(false)
				c.setHFlag(false)
				c.setCFlag(true)

				c.PC += 1
				c.waitCycles += 4
			},
			toString: "SCF",
		}, nil
	case 0x38: // JR C,r8
		return OpCode{
			execution: func(c *Cpu) {
				jumpRelative(c, c.getCFlag())
			},
			toString: "JR C,r8",
		}, nil
	case 0x39: // ADD HL,SP
		return OpCode{
			execution: func(c *Cpu) {
				addHL(c, c.SP)
			},
			toString: "ADD HL,SP",
		}, nil
	case 0x3A: // LD A,(HL-)
		return OpCode{
			execution: func(c *Cpu) {
				c.AF.upper.value = c.readFromBus(c.HL.getAll())
				c.HL.setAll(c.HL.getAll() - 1)
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD A,(HL-)",
		}, nil
	case 0x3B: // DEC SP
		return OpCode{
			execution: func(c *Cpu) {
				c.SP -= 1

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "DEC SP",
		}, nil
	case 0x3C: // INC A
		return OpCode{
			execution: func(c *Cpu) {
				inc8(c, c.AF.upper)
			},
			toString: "INC A",
		}, nil
	case 0x3D: // DEC A
		return OpCode{
			execution: func(c *Cpu) {
				dec8(c, c.AF.upper)
			},
			toString: "DEC A",
		}, nil
	case 0x3E: // LD A, u8
		return OpCode{
			execution: func(c *Cpu) {
				c.AF.upper.value = c.pcReadNext()
				c.PC += 2
				c.waitCycles += 8
			},
			toString: "LD A, u8",
		}, nil
	case 0x3F: // CCF
		return OpCode{
			execution: func(c *Cpu) {
				c.setNFlag(false)
				c.setHFlag(false)
				c.setCFlag(!c.getCFlag())
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "CCF",
		}, nil
	case 0x40: // LD B,B
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.upper, c.BC.upper)
				//fmt.Println("Executed LD B,B")
			},
			toString: "LD B,B",
		}, nil
	case 0x41: // LD B,C
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.upper, c.BC.lower)
			},
			toString: "LD B,C",
		}, nil
	case 0x42: // LD B,D
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.upper, c.DE.upper)
			},
		}, nil
	case 0x43: // LD B,E
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.upper, c.DE.lower)
			},
			toString: "LD B,E",
		}, nil
	case 0x44: // LD B,H
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.upper, c.HL.upper)
			},
			toString: "LD B,H",
		}, nil
	case 0x45: // LD B,L
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.upper, c.HL.lower)
			},
			toString: "LD B,L",
		}, nil
	case 0x46: // LD B,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				c.BC.upper.value = c.readFromBus(c.HL.getAll())
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD B,(HL)",
		}, nil
	case 0x47: // LD B,A
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.upper, c.AF.upper)
			},
			toString: "LD B,A",
		}, nil
	case 0x48: // LD C,B
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.lower, c.BC.upper)
			},
			toString: "LD C,B",
		}, nil
	case 0x49: // LD C,C
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.lower, c.BC.lower)
			},
			toString: "LD C,C",
		}, nil
	case 0x4A: // LD C,D
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.lower, c.DE.upper)
			},
			toString: "LD C,D",
		}, nil
	case 0x4B: // LD C,E
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.lower, c.DE.lower)
			},
			toString: "LD C,E",
		}, nil
	case 0x4C: // LD C,H
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.lower, c.HL.upper)
			},
			toString: "LD C,H",
		}, nil
	case 0x4D: // LD C,L
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.lower, c.HL.lower)
			},
			toString: "LD C,L",
		}, nil
	case 0x4E: // LD C,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				c.BC.lower.value = c.readFromBus(c.HL.getAll())
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD C,(HL)",
		}, nil
	case 0x4F: // LD C,A
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.BC.lower, c.AF.upper)
			},
			toString: "LD C,A",
		}, nil
	case 0x50: // LD D,B
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.upper, c.BC.upper)
			},
			toString: "LD D,B",
		}, nil
	case 0x51: // LD D,C
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.upper, c.BC.lower)
			},
			toString: "LD D,C",
		}, nil
	case 0x52: // LD D,D
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.upper, c.DE.upper)
			},
			toString: "LD D,D",
		}, nil
	case 0x53: // LD D,E
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.upper, c.DE.lower)
			},
			toString: "LD D,E",
		}, nil
	case 0x54: // LD D,H
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.upper, c.HL.upper)
			},
			toString: "LD D,H",
		}, nil
	case 0x55: // LD D,L
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.upper, c.HL.lower)
			},
			toString: "LD D,L",
		}, nil
	case 0x56: // LD D,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				c.DE.upper.value = c.readFromBus(c.HL.getAll())
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD D,(HL)",
		}, nil
	case 0x57: // LD D,A
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.upper, c.AF.upper)
			},
			toString: "LD D,A",
		}, nil
	case 0x58: // LD E,B
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.lower, c.BC.upper)
			},
			toString: "LD E,B",
		}, nil
	case 0x59: // LD E,C
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.lower, c.BC.lower)
			},
			toString: "LD E,C",
		}, nil
	case 0x5A: // LD E,D
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.lower, c.DE.upper)
			},
			toString: "LD E,D",
		}, nil
	case 0x5B: // LD E,E
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.lower, c.DE.lower)
			},
			toString: "LD E,E",
		}, nil
	case 0x5C: // LD E,H
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.lower, c.HL.upper)
			},
			toString: "LD E,H",
		}, nil
	case 0x5D: // LD E,L
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.lower, c.HL.lower)
			},
			toString: "LD E,L",
		}, nil
	case 0x5E: // LD E,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				c.DE.lower.value = c.readFromBus(c.HL.getAll())
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD E,(HL)",
		}, nil
	case 0x5F: // LD E,A
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.DE.lower, c.AF.upper)
			},
			toString: "LD E,A",
		}, nil
	case 0x60: // LD H,B
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.upper, c.BC.upper)
			},
			toString: "LD H,B",
		}, nil
	case 0x61: // LD H,C
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.upper, c.BC.lower)
			},
			toString: "LD H,C",
		}, nil
	case 0x62: // LD H,D
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.upper, c.DE.upper)
			},
			toString: "LD H,D",
		}, nil
	case 0x63: // LD H,E
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.upper, c.DE.lower)
			},
			toString: "LD H,E",
		}, nil
	case 0x64: // LD H,H
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.upper, c.HL.upper)
			},
			toString: "LD H,H",
		}, nil
	case 0x65: // LD H,L
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.upper, c.HL.lower)
			},
			toString: "LD H,L",
		}, nil
	case 0x66: // LD H,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				c.HL.upper.value = c.readFromBus(c.HL.getAll())
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD H,(HL)",
		}, nil
	case 0x67: // LD H,A
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.upper, c.AF.upper)
			},
			toString: "LD H,A",
		}, nil
	case 0x68: // LD L,B
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.lower, c.BC.upper)
			},
			toString: "LD L,B",
		}, nil
	case 0x69: // LD L,C
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.lower, c.BC.lower)
			},
			toString: "LD L,C",
		}, nil
	case 0x6A: // LD L,D
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.lower, c.DE.upper)
			},
			toString: "LD L,D",
		}, nil
	case 0x6B: // LD L,E
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.lower, c.DE.lower)
			},
			toString: "LD L,E",
		}, nil
	case 0x6C: // LD L,H
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.lower, c.HL.upper)
			},
			toString: "LD L,H",
		}, nil
	case 0x6D: // LD L,L
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.lower, c.HL.lower)
			},
			toString: "LD L,L",
		}, nil
	case 0x6E: // LD L,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				c.HL.lower.value = c.readFromBus(c.HL.getAll())
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD L,(HL)",
		}, nil
	case 0x6F: // LD L,A
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.HL.lower, c.AF.upper)
			},
			toString: "LD L,A",
		}, nil
	case 0x70: // LD (HL),B
		return OpCode{
			execution: func(c *Cpu) {
				loadAddrAtHL(c, c.BC.upper)
			},
			toString: "LD (HL),B",
		}, nil
	case 0x71: // LD (HL),C
		return OpCode{
			execution: func(c *Cpu) {
				loadAddrAtHL(c, c.BC.lower)
			},
			toString: "LD (HL),C",
		}, nil
	case 0x72: // LD (HL),D
		return OpCode{
			execution: func(c *Cpu) {
				loadAddrAtHL(c, c.DE.upper)
			},
			toString: "LD (HL),D",
		}, nil
	case 0x73: // LD (HL),E
		return OpCode{
			execution: func(c *Cpu) {
				loadAddrAtHL(c, c.DE.lower)
			},
			toString: "LD (HL),E",
		}, nil
	case 0x74: // LD (HL),H
		return OpCode{
			execution: func(c *Cpu) {
				loadAddrAtHL(c, c.HL.upper)
			},
			toString: "LD (HL),H",
		}, nil
	case 0x75: // LD (HL),L
		return OpCode{
			execution: func(c *Cpu) {
				loadAddrAtHL(c, c.HL.lower)
			},
			toString: "LD (HL),L",
		}, nil
	case 0x76: // HALT
		return OpCode{
			execution: func(c *Cpu) {
				c.halt = true
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "HALT",
		}, nil
	case 0x77: // LD (HL),A
		return OpCode{
			execution: func(c *Cpu) {
				loadAddrAtHL(c, c.AF.upper)
			},
			toString: "LD (HL),A",
		}, nil
	case 0x78: // LD A,B
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.AF.upper, c.BC.upper)
			},
			toString: "LD A,B",
		}, nil
	case 0x79: // LD A,C
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.AF.upper, c.BC.lower)
			},
			toString: "LD A,C",
		}, nil
	case 0x7A: // LD A,D
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.AF.upper, c.DE.upper)
			},
			toString: "LD A,D",
		}, nil
	case 0x7B: // LD A,E
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.AF.upper, c.DE.lower)
			},
			toString: "LD A,E",
		}, nil
	case 0x7C: // LD A,H
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.AF.upper, c.HL.upper)
			},
			toString: "LD A,H",
		}, nil
	case 0x7D: // LD A,L
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.AF.upper, c.HL.lower)
			},
			toString: "LD A,L",
		}, nil
	case 0x7E: // LD A,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				c.AF.upper.value = c.readFromBus(c.HL.getAll())
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD A,(HL)",
		}, nil
	case 0x7F: // LD A,A
		return OpCode{
			execution: func(c *Cpu) {
				loadRegister(c, c.AF.upper, c.AF.upper)
			},
			toString: "LD A,A",
		}, nil
	case 0x80: // ADD A,B
		return OpCode{
			execution: func(c *Cpu) {
				add(c, c.BC.upper.value)
			},
			toString: "ADD A,B",
		}, nil
	case 0x81: // ADD A,C
		return OpCode{
			execution: func(c *Cpu) {
				add(c, c.BC.lower.value)
			},
			toString: "ADD A,C",
		}, nil
	case 0x82: // ADD A,D
		return OpCode{
			execution: func(c *Cpu) {
				add(c, c.DE.upper.value)
			},
			toString: "ADD A,D",
		}, nil
	case 0x83: // ADD A,E
		return OpCode{
			execution: func(c *Cpu) {
				add(c, c.DE.lower.value)
			},
			toString: "ADD A,E",
		}, nil
	case 0x84: // ADD A,H
		return OpCode{
			execution: func(c *Cpu) {
				add(c, c.HL.upper.value)
			},
			toString: "ADD A,H",
		}, nil
	case 0x85: // ADD A,L
		return OpCode{
			execution: func(c *Cpu) {
				add(c, c.HL.lower.value)
			},
			toString: "ADD A,L",
		}, nil
	case 0x86: // ADD A,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				orgVal := c.AF.upper.value
				addVal := c.readFromBus(c.HL.getAll())

				c.setZFlag(orgVal+addVal == 0)
				c.setNFlag(false)
				c.setHFlag((orgVal&0x0F)+(addVal&0x0F) > 0x0F)
				c.setCFlag(uint16(orgVal)+uint16(addVal) > 0xFF)

				c.AF.upper.value += addVal

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "ADD A,(HL)",
		}, nil
	case 0x87: // ADD A,A
		return OpCode{
			execution: func(c *Cpu) {
				add(c, c.AF.upper.value)
			},
			toString: "ADD A,A",
		}, nil
	case 0x88: // ADC A,B
		return OpCode{
			execution: func(c *Cpu) {
				adc(c, c.BC.upper.value)
			},
			toString: "ADC A,B",
		}, nil
	case 0x89: // ADC A,C
		return OpCode{
			execution: func(c *Cpu) {
				adc(c, c.BC.lower.value)
			},
			toString: "ADC A,C",
		}, nil
	case 0x8A: // ADC A,D
		return OpCode{
			execution: func(c *Cpu) {
				adc(c, c.DE.upper.value)
			},
			toString: "ADC A,D",
		}, nil
	case 0x8B: // ADC A,E
		return OpCode{
			execution: func(c *Cpu) {
				adc(c, c.DE.lower.value)
			},
			toString: "ADC A,E",
		}, nil
	case 0x8C: // ADC A,H
		return OpCode{
			execution: func(c *Cpu) {
				adc(c, c.HL.upper.value)
			},
			toString: "ADC A,H",
		}, nil
	case 0x8D: // ADC A,L
		return OpCode{
			execution: func(c *Cpu) {
				adc(c, c.HL.lower.value)
			},
			toString: "ADC A,L",
		}, nil
	case 0x8E: // ADC A,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				orgVal := c.AF.upper.value
				addVal := c.readFromBus(c.HL.getAll())

				var carryVal byte = 0
				if c.getCFlag() {
					carryVal = 1
				}

				c.setZFlag((orgVal + addVal + carryVal) == 0)
				c.setNFlag(false)
				c.setHFlag((orgVal&0x0F)+(addVal&0x0F)+(carryVal&0x0F) > 0x0F)
				c.setCFlag(uint16(orgVal)+uint16(addVal)+uint16(carryVal) > 0xFF)

				c.AF.upper.value += addVal + carryVal

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "ADC A,(HL)",
		}, nil
	case 0x8F: // ADC A,A
		return OpCode{
			execution: func(c *Cpu) {
				adc(c, c.AF.upper.value)
			},
			toString: "ADC A,A",
		}, nil
	case 0x90: // SUB B
		return OpCode{
			execution: func(c *Cpu) {
				sub(c, c.BC.upper.value)
			},
			toString: "SUB B",
		}, nil
	case 0x91: // SUB C
		return OpCode{
			execution: func(c *Cpu) {
				sub(c, c.BC.lower.value)
			},
			toString: "SUB C",
		}, nil
	case 0x92: // SUB D
		return OpCode{
			execution: func(c *Cpu) {
				sub(c, c.DE.upper.value)
			},
			toString: "SUB D",
		}, nil
	case 0x93: // SUB E
		return OpCode{
			execution: func(c *Cpu) {
				sub(c, c.DE.lower.value)
			},
			toString: "SUB E",
		}, nil
	case 0x94: // SUB H
		return OpCode{
			execution: func(c *Cpu) {
				sub(c, c.HL.upper.value)
			},
			toString: "SUB H",
		}, nil
	case 0x95: // SUB L
		return OpCode{
			execution: func(c *Cpu) {
				sub(c, c.HL.lower.value)
			},
			toString: "SUB L",
		}, nil
	case 0x96: // SUB A,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				orgVal := c.AF.upper.value
				subVal := c.readFromBus(c.HL.getAll())
				c.AF.upper.value -= subVal

				c.setZFlag(c.AF.upper.value == 0)
				c.setNFlag(true)
				c.setHFlag(int8((orgVal&0xF)-(subVal&0xF)) < 0)
				c.setCFlag(orgVal < subVal)

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "SUB A,(HL)",
		}, nil
	case 0x97: // SUB A
		return OpCode{
			execution: func(c *Cpu) {
				sub(c, c.AF.upper.value)
			},
			toString: "SUB A",
		}, nil
	case 0x98: // SBC A,B
		return OpCode{
			execution: func(c *Cpu) {
				sbc(c, c.BC.upper.value)
			},
			toString: "SBC A,B",
		}, nil
	case 0x99: // SBC A,C
		return OpCode{
			execution: func(c *Cpu) {
				sbc(c, c.BC.lower.value)
			},
			toString: "SBC A,C",
		}, nil
	case 0x9A: // SBC A,D
		return OpCode{
			execution: func(c *Cpu) {
				sbc(c, c.DE.upper.value)
			},
			toString: "SBC A,D",
		}, nil
	case 0x9B: // SBC A,E
		return OpCode{
			execution: func(c *Cpu) {
				sbc(c, c.DE.lower.value)
			},
			toString: "SBC A,E",
		}, nil
	case 0x9C: // SBC A,H
		return OpCode{
			execution: func(c *Cpu) {
				sbc(c, c.HL.upper.value)
			},
			toString: "SBC A,H",
		}, nil
	case 0x9D: // SBC A,L
		return OpCode{
			execution: func(c *Cpu) {
				sbc(c, c.HL.lower.value)
			},
			toString: "SBC A,L",
		}, nil
	case 0x9E: // SBC A,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				orgVal := c.AF.upper.value
				subVal := c.readFromBus(c.HL.getAll())
				var carryVal byte = 0
				if c.getCFlag() {
					carryVal = 1
				}

				c.setCFlag(int32(orgVal)-int32(subVal)-int32(carryVal) < 0)
				c.setHFlag((int32(orgVal&0x0F) - int32(subVal&0x0F) - int32(carryVal)) < 0)
				c.setNFlag(true)
				c.setZFlag((orgVal - subVal - carryVal) == 0)

				c.AF.upper.value = c.AF.upper.value - subVal - carryVal

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "SBC A,(HL)",
		}, nil
	case 0x9F: // SBC A,A
		return OpCode{
			execution: func(c *Cpu) {
				sbc(c, c.AF.upper.value)
			},
			toString: "SBC A,A",
		}, nil
	case 0xA0: // AND B
		return OpCode{
			execution: func(c *Cpu) {
				and(c, c.BC.upper.value)
			},
			toString: "AND B",
		}, nil
	case 0xA1: // AND C
		return OpCode{
			execution: func(c *Cpu) {
				and(c, c.BC.lower.value)
			},
			toString: "AND C",
		}, nil
	case 0xA2: // AND D
		return OpCode{
			execution: func(c *Cpu) {
				and(c, c.DE.upper.value)
			},
			toString: "AND D",
		}, nil
	case 0xA3: // AND E
		return OpCode{
			execution: func(c *Cpu) {
				and(c, c.DE.lower.value)
			},
			toString: "AND E",
		}, nil
	case 0xA4: // AND H
		return OpCode{
			execution: func(c *Cpu) {
				and(c, c.HL.upper.value)
			},
			toString: "AND H",
		}, nil
	case 0xA5: // AND L
		return OpCode{
			execution: func(c *Cpu) {
				and(c, c.HL.lower.value)
			},
			toString: "AND L",
		}, nil
	case 0xA6: // AND (HL)
		return OpCode{
			execution: func(c *Cpu) {
				c.AF.upper.value &= c.readFromBus(c.HL.getAll())

				c.setZFlag(c.AF.upper.value == 0)
				c.setNFlag(false)
				c.setHFlag(true)
				c.setCFlag(false)

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "AND (HL)",
		}, nil
	case 0xA7: // AND A
		return OpCode{
			execution: func(c *Cpu) {
				and(c, c.AF.upper.value)
			},
			toString: "AND A",
		}, nil
	case 0xA8: // XOR B
		return OpCode{
			execution: func(c *Cpu) {
				xor(c, c.BC.upper.value)
			},
			toString: "XOR B",
		}, nil
	case 0xA9: // XOR C
		return OpCode{
			execution: func(c *Cpu) {
				xor(c, c.BC.lower.value)
			},
			toString: "XOR C",
		}, nil
	case 0xAA: // XOR D
		return OpCode{
			execution: func(c *Cpu) {
				xor(c, c.DE.upper.value)
			},
			toString: "XOR D",
		}, nil
	case 0xAB: // XOR E
		return OpCode{
			execution: func(c *Cpu) {
				xor(c, c.DE.lower.value)
			},
			toString: "XOR E",
		}, nil
	case 0xAC: // XOR H
		return OpCode{
			execution: func(c *Cpu) {
				xor(c, c.HL.upper.value)
			},
			toString: "XOR H",
		}, nil
	case 0xAD: // XOR L
		return OpCode{
			execution: func(c *Cpu) {
				xor(c, c.HL.lower.value)
			},
			toString: "XOR L",
		}, nil
	case 0xAE: // XOR (HL)
		return OpCode{
			execution: func(c *Cpu) {
				c.AF.upper.value ^= c.readFromBus(c.HL.getAll())

				c.setZFlag(c.AF.upper.value == 0)
				c.setNFlag(false)
				c.setHFlag(false)
				c.setCFlag(false)

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "XOR (HL)",
		}, nil
	case 0xAF: // XOR A
		return OpCode{
			execution: func(c *Cpu) {
				xor(c, c.AF.upper.value)
			},
			toString: "XOR A",
		}, nil
	case 0xB0: // OR B
		return OpCode{
			execution: func(c *Cpu) {
				or(c, c.BC.upper.value)
			},
			toString: "OR B",
		}, nil
	case 0xB1: // OR C
		return OpCode{
			execution: func(c *Cpu) {
				or(c, c.BC.lower.value)
			},
			toString: "OR C",
		}, nil
	case 0xB2: // OR D
		return OpCode{
			execution: func(c *Cpu) {
				or(c, c.DE.upper.value)
			},
			toString: "OR D",
		}, nil
	case 0xB3: // OR E
		return OpCode{
			execution: func(c *Cpu) {
				or(c, c.DE.lower.value)
			},
			toString: "OR E",
		}, nil
	case 0xB4: // OR H
		return OpCode{
			execution: func(c *Cpu) {
				or(c, c.HL.upper.value)
			},
			toString: "OR H",
		}, nil
	case 0xB5: // OR L
		return OpCode{
			execution: func(c *Cpu) {
				or(c, c.HL.lower.value)
			},
			toString: "OR L",
		}, nil
	case 0xB6: // OR (HL)
		return OpCode{
			execution: func(c *Cpu) {
				c.AF.upper.value |= c.readFromBus(c.HL.getAll())

				c.setZFlag(c.AF.upper.value == 0)
				c.setNFlag(false)
				c.setHFlag(false)
				c.setCFlag(false)

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "OR (HL)",
		}, nil
	case 0xB7: // OR A
		return OpCode{
			execution: func(c *Cpu) {
				or(c, c.AF.upper.value)
			},
			toString: "OR A",
		}, nil
	case 0xB8: // CP B
		return OpCode{
			execution: func(c *Cpu) {
				cp(c, c.BC.upper.value)
			},
			toString: "CP B",
		}, nil
	case 0xB9: // CP C
		return OpCode{
			execution: func(c *Cpu) {
				cp(c, c.BC.lower.value)
			},
			toString: "CP C",
		}, nil
	case 0xBA: // CP D
		return OpCode{
			execution: func(c *Cpu) {
				cp(c, c.DE.upper.value)
			},
			toString: "CP D",
		}, nil
	case 0xBB: // CP E
		return OpCode{
			execution: func(c *Cpu) {
				cp(c, c.DE.lower.value)
			},
			toString: "CP E",
		}, nil
	case 0xBC: // CP H
		return OpCode{
			execution: func(c *Cpu) {
				cp(c, c.HL.upper.value)
			},
			toString: "CP H",
		}, nil
	case 0xBD: // CP L
		return OpCode{
			execution: func(c *Cpu) {
				cp(c, c.HL.lower.value)
			},
			toString: "CP L",
		}, nil
	case 0xBE: // CP (HL)
		return OpCode{
			execution: func(c *Cpu) {
				hlVal := c.readFromBus(c.HL.getAll())

				c.setZFlag(c.AF.upper.value-hlVal == 0)
				c.setNFlag(true)
				c.setHFlag((c.AF.upper.value & 0x0F) < (hlVal & 0x0F))
				c.setCFlag(c.AF.upper.value < hlVal)

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "CP (HL)",
		}, nil
	case 0xBF: // CP A
		return OpCode{
			execution: func(c *Cpu) {
				cp(c, c.AF.upper.value)
			},
			toString: "CP A",
		}, nil
	case 0xC0: // RET NZ
		return OpCode{
			execution: func(c *Cpu) {
				ret(c, !c.getZFlag())
			},
			toString: "RET NZ",
		}, nil
	case 0xC1: // POP BC
		return OpCode{
			execution: func(c *Cpu) {
				pop(c, c.BC)
			},
			toString: "POP BC",
		}, nil
	case 0xC2: // JP NZ,u16
		return OpCode{
			execution: func(c *Cpu) {
				jump(c, !c.getZFlag())
			},
			toString: "JP NZ,u16",
		}, nil
	case 0xC3: // JP u16
		return OpCode{
			execution: func(c *Cpu) {
				jump(c, true)
			},
			toString: "JP u16",
		}, nil
	case 0xC4: // CALL NZ,u16
		return OpCode{
			execution: func(c *Cpu) {
				call(c, !c.getZFlag())
			},
			toString: "CALL NZ,u16",
		}, nil
	case 0xC5: // PUSH BC
		return OpCode{
			execution: func(c *Cpu) {
				push(c, c.BC)
			},
			toString: "PUSH BC",
		}, nil
	case 0xC6: // ADD A,u8
		return OpCode{
			execution: func(c *Cpu) {
				add(c, c.pcReadNext())
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "ADD A,u8",
		}, nil
	case 0xC7: // RST 00H
		return OpCode{
			execution: func(c *Cpu) {
				rst(c, 0x00)
			},
			toString: "RST 00H",
		}, nil
	case 0xC8: // RET Z
		return OpCode{
			execution: func(c *Cpu) {
				ret(c, c.getZFlag())
			},
			toString: "RET Z",
		}, nil
	case 0xC9: // RET
		return OpCode{
			execution: func(c *Cpu) {
				low := uint16(c.readFromBus(c.SP))
				c.SP++
				high := uint16(c.readFromBus(c.SP))
				c.SP++

				c.PC = high<<8 | low
				c.waitCycles += 16
			},
			toString: "RET",
		}, nil
	case 0xCA: // JP Z,u16
		return OpCode{
			execution: func(c *Cpu) {
				jump(c, c.getZFlag())
			},
			toString: "JP Z,u16",
		}, nil
	case 0xCC: // CALL Z,u16
		return OpCode{
			execution: func(c *Cpu) {
				call(c, c.getZFlag())
			},
			toString: "CALL Z,u16",
		}, nil
	case 0xCD: // CALL u16
		return OpCode{
			execution: func(c *Cpu) {
				call(c, true)
			},
			toString: "CALL u16",
		}, nil
	case 0xCE: // ADC A,d8
		return OpCode{
			execution: func(c *Cpu) {
				adc(c, c.pcReadNext())
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "ADC A,d8",
		}, nil
	case 0xCF: // RST 08H
		return OpCode{
			execution: func(c *Cpu) {
				rst(c, 0x08)
			},
			toString: "RST 08H",
		}, nil
	case 0xD0: // RET NC
		return OpCode{
			execution: func(c *Cpu) {
				ret(c, !c.getCFlag())
			},
			toString: "RET NC",
		}, nil
	case 0xD1: // POP DE
		return OpCode{
			execution: func(c *Cpu) {
				pop(c, c.DE)
			},
			toString: "POP DE",
		}, nil
	case 0xD2: // JP NC,u16
		return OpCode{
			execution: func(c *Cpu) {
				jump(c, !c.getCFlag())
			},
			toString: "JP NC,u16",
		}, nil
	case 0xD4: // CALL NC,u16
		return OpCode{
			execution: func(c *Cpu) {
				call(c, !c.getCFlag())
			},
			toString: "CALL NC,u16",
		}, nil
	case 0xD5: // PUSH DE
		return OpCode{
			execution: func(c *Cpu) {
				push(c, c.DE)
			},
			toString: "PUSH DE",
		}, nil
	case 0xD6: // SUB A,u8
		return OpCode{
			execution: func(c *Cpu) {
				sub(c, c.pcReadNext())
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "SUB A,u8",
		}, nil
	case 0xD7: // RST 10H
		return OpCode{
			execution: func(c *Cpu) {
				rst(c, 0x10)
			},
			toString: "RST 10H",
		}, nil
	case 0xD8: // RET C
		return OpCode{
			execution: func(c *Cpu) {
				ret(c, c.getCFlag())
			},
			toString: "RET C",
		}, nil
	case 0xD9: // RETI
		return OpCode{
			execution: func(c *Cpu) {
				low := uint16(c.readFromBus(c.SP))
				c.SP++
				high := uint16(c.readFromBus(c.SP))
				c.SP++

				c.PC = high<<8 | low
				c.waitCycles += 16
				c.interruptEnabled = true
			},
			toString: "RETI",
		}, nil
	case 0xDA: // JP C,u16
		return OpCode{
			execution: func(c *Cpu) {
				jump(c, c.getCFlag())
			},
			toString: "JP C,u16",
		}, nil
	case 0xDC: // CALL C,u16
		return OpCode{
			execution: func(c *Cpu) {
				call(c, c.getCFlag())
			},
			toString: "CALL C,u16",
		}, nil
	case 0xDE: // SBC A,d8
		return OpCode{
			execution: func(c *Cpu) {
				sbc(c, c.pcReadNext())
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "SBC A,d8",
		}, nil
	case 0xDF: // RST 18H
		return OpCode{
			execution: func(c *Cpu) {
				rst(c, 0x18)
			},
			toString: "RST 18H",
		}, nil
	case 0xE0: // LDH (a8),A
		return OpCode{
			execution: func(c *Cpu) {
				c.writeToBus(0xFF00|uint16(c.pcReadNext()), c.AF.upper.value)
				c.PC += 2
				c.waitCycles += 12
			},
			toString: "LDH (a8),A",
		}, nil
	case 0xE1: // POP HL
		return OpCode{
			execution: func(c *Cpu) {
				pop(c, c.HL)
			},
			toString: "POP HL",
		}, nil
	case 0xE2: // LDH (C),A
		return OpCode{
			execution: func(c *Cpu) {
				c.writeToBus(0xFF00|uint16(c.BC.lower.value), c.AF.upper.value)
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LDH (C),A",
		}, nil
	case 0xE5: // PUSH HL
		return OpCode{
			execution: func(c *Cpu) {
				push(c, c.HL)
			},
			toString: "PUSH HL",
		}, nil
	case 0xE6: // AND A,u8
		return OpCode{
			execution: func(c *Cpu) {
				and(c, c.pcReadNext())
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "AND A,u8",
		}, nil
	case 0xE7: // RST 20H
		return OpCode{
			execution: func(c *Cpu) {
				rst(c, 0x20)
			},
			toString: "RST 20H",
		}, nil
	case 0xE8: // ADD SP,r8
		return OpCode{
			execution: func(c *Cpu) {
				addVal := int32(int8(c.pcReadNext()))
				spVal := int32(c.SP)
				c.SP = uint16(spVal + addVal)

				c.setZFlag(false)
				c.setNFlag(false)
				c.setHFlag((spVal&0x0F)+(addVal&0x0F) > 0x0F)
				c.setCFlag((spVal&0xFF)+(addVal&0xFF) > 0xFF)

				c.PC += 2
				c.waitCycles += 16
			},
			toString: "ADD SP,r8",
		}, nil
	case 0xE9: // JP (HL)
		return OpCode{
			execution: func(c *Cpu) {
				c.PC = c.HL.getAll()
				c.waitCycles += 4
			},
			toString: "JP (HL)",
		}, nil
	case 0xEA: // LD (u16),A
		return OpCode{
			execution: func(c *Cpu) {
				c.writeToBus(c.pcReadNext16(), c.AF.upper.value)
				c.PC += 3
				c.waitCycles += 16
			},
			toString: "LD (u16),A",
		}, nil
	case 0xEE: // XOR u8
		return OpCode{
			execution: func(c *Cpu) {
				xor(c, c.pcReadNext())
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "XOR u8",
		}, nil
	case 0xEF: // RST 28H
		return OpCode{
			execution: func(c *Cpu) {
				rst(c, 0x28)
			},
			toString: "RST 28H",
		}, nil
	case 0xF0: // LDH A,(a8)
		return OpCode{
			execution: func(c *Cpu) {
				c.AF.upper.value = c.readFromBus(0xFF00 | uint16(c.pcReadNext()))
				c.PC += 2
				c.waitCycles += 12
			},
			toString: "LDH A,(a8)",
		}, nil
	case 0xF1: // POP AF
		return OpCode{
			execution: func(c *Cpu) {
				pop(c, c.AF)
				c.AF.lower.value &= 0xF0
			},
			toString: "POP AF",
		}, nil
	case 0xF2: // LDH A,(C)
		return OpCode{
			execution: func(c *Cpu) {
				c.AF.upper.value = c.readFromBus(0xFF00 | uint16(c.BC.lower.value))
				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LDH A,(C)",
		}, nil
	case 0xF3: // DI
		return OpCode{
			execution: func(c *Cpu) {
				c.interruptEnabled = false
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "DI",
		}, nil
	case 0xF5: // PUSH AF
		return OpCode{
			execution: func(c *Cpu) {
				push(c, c.AF)
			},
			toString: "PUSH AF",
		}, nil
	case 0xF6: // OR A,u8
		return OpCode{
			execution: func(c *Cpu) {
				or(c, c.pcReadNext())
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "OR A,u8",
		}, nil
	case 0xF7: // RST 30H
		return OpCode{
			execution: func(c *Cpu) {
				rst(c, 0x30)
			},
			toString: "RST 30H",
		}, nil
	case 0xF8: // LD HL,SP+r8
		return OpCode{
			execution: func(c *Cpu) {
				addVal := uint16(int8(c.pcReadNext()))

				c.setZFlag(false)
				c.setNFlag(false)
				c.setHFlag((c.SP&0x0F)+(addVal&0x0F) > 0x0F)
				c.setCFlag((c.SP&0xFF)+(addVal&0xFF) > 0xFF)

				c.HL.setAll(c.SP + addVal)

				c.PC += 2
				c.waitCycles += 12
			},
			toString: "LD HL,SP+r8",
		}, nil
	case 0xF9: // LD SP,HL
		return OpCode{
			execution: func(c *Cpu) {
				c.SP = c.HL.getAll()

				c.PC += 1
				c.waitCycles += 8
			},
			toString: "LD SP,HL",
		}, nil
	case 0xFA: // LD A,(u16)
		return OpCode{
			execution: func(c *Cpu) {
				c.AF.upper.value = c.readFromBus(c.pcReadNext16())

				c.PC += 3
				c.waitCycles += 16
			},
			toString: "LD A,(u16)",
		}, nil
	case 0xFB: // EI
		return OpCode{
			execution: func(c *Cpu) {
				c.interruptEnabled = true

				c.PC += 1
				c.waitCycles += 4
			},
			toString: "EI",
		}, nil
	case 0xFE: // CP u8
		return OpCode{
			execution: func(c *Cpu) {
				cp(c, c.pcReadNext())
				c.PC += 1
				c.waitCycles += 4
			},
			toString: "CP u8",
		}, nil
	case 0xFF: // RST 38H
		return OpCode{
			execution: func(c *Cpu) {
				rst(c, 0x38)
			},
			toString: "RST 38H",
		}, nil
	default:
		return OpCode{}, fmt.Errorf("opcode %d not recognized", code)
	}
}

func rst(c *Cpu, pcVal uint16) {
	c.PC += 1
	c.SP--
	c.writeToBus(c.SP, byte(c.PC>>8))
	c.SP--
	c.writeToBus(c.SP, byte(c.PC&0xFF))
	c.PC = pcVal
	c.waitCycles += 16
}

func call(c *Cpu, condition bool) {
	if condition {
		retVal := c.PC + 3
		c.SP--
		c.writeToBus(c.SP, byte(retVal>>8))
		c.SP--
		c.writeToBus(c.SP, byte(retVal&0xFF))

		c.PC = c.pcReadNext16()
		c.waitCycles += 24
	} else {
		c.PC += 3
		c.waitCycles += 12
	}
}

func push(c *Cpu, register *Register) {
	c.SP--
	c.writeToBus(c.SP, register.upper.value)
	c.SP--
	c.writeToBus(c.SP, register.lower.value)
	c.PC += 1
	c.waitCycles += 16
}

func pop(c *Cpu, register *Register) {
	register.lower.value = c.readFromBus(c.SP)
	c.SP++
	register.upper.value = c.readFromBus(c.SP)
	c.SP++
	c.PC += 1
	c.waitCycles += 12
}

func ret(c *Cpu, condition bool) {
	if condition {
		low := uint16(c.readFromBus(c.SP))
		c.SP++
		high := uint16(c.readFromBus(c.SP))
		c.SP++

		c.PC = high<<8 | low
		c.waitCycles += 20
	} else {
		c.PC += 1
		c.waitCycles += 8
	}
}

func cp(c *Cpu, operand byte) {
	c.setZFlag(c.AF.upper.value == operand)
	c.setNFlag(true)
	c.setHFlag((c.AF.upper.value & 0x0F) < (operand & 0x0F))
	c.setCFlag(c.AF.upper.value < operand)

	c.PC += 1
	c.waitCycles += 4
}

func or(c *Cpu, operand byte) {
	c.AF.upper.value |= operand

	c.setZFlag(c.AF.upper.value == 0)
	c.setNFlag(false)
	c.setHFlag(false)
	c.setCFlag(false)

	c.PC += 1
	c.waitCycles += 4
}

func xor(c *Cpu, operand byte) {
	c.AF.upper.value ^= operand

	c.setZFlag(c.AF.upper.value == 0)
	c.setNFlag(false)
	c.setHFlag(false)
	c.setCFlag(false)

	c.PC += 1
	c.waitCycles += 4
}

func and(c *Cpu, operand byte) {
	c.AF.upper.value &= operand

	c.setZFlag(c.AF.upper.value == 0)
	c.setNFlag(false)
	c.setHFlag(true)
	c.setCFlag(false)

	c.PC += 1
	c.waitCycles += 4
}

func sbc(c *Cpu, operand byte) {
	var carryVal byte = 0
	if c.getCFlag() {
		carryVal = 1
	}

	c.setCFlag(int32(c.AF.upper.value)-int32(operand)-int32(carryVal) < 0)
	c.setHFlag((int32(c.AF.upper.value&0x0F) - int32(operand&0x0F) - int32(carryVal)) < 0)
	c.setNFlag(true)
	c.setZFlag((c.AF.upper.value - operand - carryVal) == 0)

	c.AF.upper.value = c.AF.upper.value - operand - carryVal

	c.PC += 1
	c.waitCycles += 4
}

func sub(c *Cpu, operand byte) {
	c.setZFlag((c.AF.upper.value - operand) == 0)
	c.setNFlag(true)
	c.setHFlag((c.AF.upper.value & 0xF) < (operand & 0xF))
	c.setCFlag(uint16(c.AF.upper.value) < uint16(operand))

	c.AF.upper.value -= operand

	c.PC += 1
	c.waitCycles += 4
}

func adc(c *Cpu, operand byte) {
	var carryVal byte = 0
	if c.getCFlag() {
		carryVal = 1
	}

	c.setZFlag((c.AF.upper.value + operand + carryVal) == 0)
	c.setNFlag(false)
	c.setHFlag((c.AF.upper.value&0x0F)+(operand&0x0F)+(carryVal&0x0F) > 0x0F)
	c.setCFlag(uint16(c.AF.upper.value)+uint16(operand)+uint16(carryVal) > 0xFF)

	c.AF.upper.value += operand + carryVal

	c.PC += 1
	c.waitCycles += 4
}

func add(c *Cpu, operand byte) {
	c.setZFlag((c.AF.upper.value + operand) == 0)
	c.setNFlag(false)
	c.setHFlag((c.AF.upper.value&0x0F)+(operand&0x0F) > 0x0F)
	c.setCFlag(uint16(c.AF.upper.value)+uint16(operand) > 0xFF)

	c.AF.upper.value += operand

	c.PC += 1
	c.waitCycles += 4
}

func loadAddrAtHL(c *Cpu, halfRegister *HalfRegister) {
	c.writeToBus(c.HL.getAll(), halfRegister.value)
	c.PC += 1
	c.waitCycles += 8
}

func loadRegister(c *Cpu, destReg *HalfRegister, srcReg *HalfRegister) {
	destReg.value = srcReg.value
	c.PC += 1
	c.waitCycles += 4
}

func jump(c *Cpu, condition bool) {
	if condition {
		c.PC = c.pcReadNext16()
		c.waitCycles += 16
	} else {
		c.PC += 3
		c.waitCycles += 12
	}
}

func jumpRelative(c *Cpu, condition bool) {
	if condition {
		jumpVal := int8(c.pcReadNext())
		c.PC = uint16(int32(c.PC) + 2 + int32(jumpVal))
		c.waitCycles += 12
	} else {
		c.PC += 2
		c.waitCycles += 8
	}
}

func addHL(c *Cpu, val uint16) {
	c.setNFlag(false)
	c.setCFlag(c.HL.getAll() > 0xFFFF-val)
	c.setHFlag(((c.HL.getAll() & 0x0FFF) + (val & 0x0FFF)) > 0x0FFF)

	c.HL.setAll(c.HL.getAll() + val)
	c.PC += 1
	c.waitCycles += 8
}

func load(r *HalfRegister, val uint8) {
	r.value = val
}

func inc8(cpu *Cpu, halfRegister *HalfRegister) {
	orgVal := halfRegister.value
	halfRegister.value++

	cpu.setZFlag(halfRegister.value == 0)
	cpu.setNFlag(false)
	cpu.setHFlag((orgVal&0x0F)+0x01 > 0x0F)

	cpu.PC += 1
	cpu.waitCycles += 4
}

func dec8(cpu *Cpu, halfRegister *HalfRegister) {
	halfRegister.value--

	cpu.setZFlag(halfRegister.value == 0)
	cpu.setNFlag(true)
	cpu.setHFlag((halfRegister.value & 0x0F) == 0x0F)

	cpu.PC += 1
	cpu.waitCycles += 4
}
