package cpu

import "fmt"

func GetOpCodeCB(code byte) (OpCode, error) {
	switch code {
	case 0x00: // RLC B
		return OpCode{
			execution: func(c *Cpu) {
				rlc(c, c.BC.upper)
			},
			toString: "RLC B",
		}, nil
	case 0x01: // RLC C
		return OpCode{
			execution: func(c *Cpu) {
				rlc(c, c.BC.lower)
			},
			toString: "RLC C",
		}, nil
	case 0x02: // RLC D
		return OpCode{
			execution: func(c *Cpu) {
				rlc(c, c.DE.upper)
			},
			toString: "RLC D",
		}, nil
	case 0x03: // RLC E
		return OpCode{
			execution: func(c *Cpu) {
				rlc(c, c.DE.lower)
			},
			toString: "RLC E",
		}, nil
	case 0x04: // RLC H
		return OpCode{
			execution: func(c *Cpu) {
				rlc(c, c.HL.upper)
			},
			toString: "RLC H",
		}, nil
	case 0x05: // RLC L
		return OpCode{
			execution: func(c *Cpu) {
				rlc(c, c.HL.lower)
			},
			toString: "RLC L",
		}, nil
	case 0x06: // RLC (HL)
		return OpCode{
			execution: func(c *Cpu) {
				orgVal := c.readFromBus(c.HL.getAll())
				newVal := orgVal
				newVal <<= 1

				if orgVal&0x80 == 0x80 {
					newVal |= 1
					c.setCFlag(true)
				} else {
					c.setCFlag(false)
				}

				c.setNFlag(false)
				c.setZFlag(newVal == 0)
				c.setHFlag(false)

				c.writeToBus(c.HL.getAll(), newVal)

				c.PC += 2
				c.waitCycles += 16
			},
			toString: "RLC (HL)",
		}, nil
	case 0x07: // RLC A
		return OpCode{
			execution: func(c *Cpu) {
				rlc(c, c.AF.upper)
			},
			toString: "RLC A",
		}, nil
	case 0x08: // RRC B
		return OpCode{
			execution: func(c *Cpu) {
				rrc(c, c.BC.upper)
			},
			toString: "RRC B",
		}, nil
	case 0x09: // RRC C
		return OpCode{
			execution: func(c *Cpu) {
				rrc(c, c.BC.lower)
			},
			toString: "RRC C",
		}, nil
	case 0x0A: // RRC D
		return OpCode{
			execution: func(c *Cpu) {
				rrc(c, c.DE.upper)
			},
			toString: "RRC D",
		}, nil
	case 0x0B: // RRC E
		return OpCode{
			execution: func(c *Cpu) {
				rrc(c, c.DE.lower)
			},
			toString: "RRC E",
		}, nil
	case 0x0C: // RRC H
		return OpCode{
			execution: func(c *Cpu) {
				rrc(c, c.HL.upper)
			},
			toString: "RRC H",
		}, nil
	case 0x0D: // RRC L
		return OpCode{
			execution: func(c *Cpu) {
				rrc(c, c.HL.lower)
			},
			toString: "RRC L",
		}, nil
	case 0x0E: // RRC (HL)
		return OpCode{
			execution: func(c *Cpu) {
				orgVal := c.readFromBus(c.HL.getAll())
				newVal := orgVal
				newVal >>= 1

				if orgVal&1 == 1 {
					newVal |= 0x80
					c.setCFlag(true)
				} else {
					c.setCFlag(false)
				}

				c.setNFlag(false)
				c.setZFlag(newVal == 0)
				c.setHFlag(false)

				c.writeToBus(c.HL.getAll(), newVal)

				c.PC += 2
				c.waitCycles += 16
			},
			toString: "RRC (HL)",
		}, nil
	case 0x0F: // RRC A
		return OpCode{
			execution: func(c *Cpu) {
				rrc(c, c.AF.upper)
			},
			toString: "RRC A",
		}, nil
	case 0x10: // RL B
		return OpCode{
			execution: func(c *Cpu) {
				rl(c, c.BC.upper)
			},
			toString: "RL B",
		}, nil
	case 0x11: // RL C
		return OpCode{
			execution: func(c *Cpu) {
				rl(c, c.BC.lower)
			},
			toString: "RL C",
		}, nil
	case 0x12: // RL D
		return OpCode{
			execution: func(c *Cpu) {
				rl(c, c.DE.upper)
			},
			toString: "RL D",
		}, nil
	case 0x13: // RL E
		return OpCode{
			execution: func(c *Cpu) {
				rl(c, c.DE.lower)
			},
			toString: "RL E",
		}, nil
	case 0x14: // RL H
		return OpCode{
			execution: func(c *Cpu) {
				rl(c, c.HL.upper)
			},
			toString: "RL H",
		}, nil
	case 0x15: // RL L
		return OpCode{
			execution: func(c *Cpu) {
				rl(c, c.HL.lower)
			},
			toString: "RL L",
		}, nil
	case 0x16: // RL (HL)
		return OpCode{
			execution: func(c *Cpu) {
				orgVal := c.readFromBus(c.HL.getAll())
				carryVal := 0
				if c.getCFlag() {
					carryVal = 1
				}
				val := orgVal<<1 | uint8(carryVal)

				c.setCFlag(orgVal&0x80 == 0x80)
				c.setNFlag(false)
				c.setZFlag(val == 0)
				c.setHFlag(false)

				c.writeToBus(c.HL.getAll(), val)

				c.PC += 2
				c.waitCycles += 16
			},
			toString: "RL (HL)",
		}, nil
	case 0x17: // RL A
		return OpCode{
			execution: func(c *Cpu) {
				rl(c, c.AF.upper)
			},
			toString: "RL A",
		}, nil
	case 0x18: // RR B
		return OpCode{
			execution: func(c *Cpu) {
				rr(c, c.BC.upper)
			},
			toString: "RR B",
		}, nil
	case 0x19: // RR C
		return OpCode{
			execution: func(c *Cpu) {
				rr(c, c.BC.lower)
			},
			toString: "RR C",
		}, nil
	case 0x1A: // RR D
		return OpCode{
			execution: func(c *Cpu) {
				rr(c, c.DE.upper)
			},
			toString: "RR D",
		}, nil
	case 0x1B: // RR E
		return OpCode{
			execution: func(c *Cpu) {
				rr(c, c.DE.lower)
			},
			toString: "RR E",
		}, nil
	case 0x1C: // RR H
		return OpCode{
			execution: func(c *Cpu) {
				rr(c, c.HL.upper)
			},
		}, nil
	case 0x1D: // RR L
		return OpCode{
			execution: func(c *Cpu) {
				rr(c, c.HL.lower)
			},
			toString: "RR L",
		}, nil
	case 0x1E: // RR (HL)
		return OpCode{
			execution: func(c *Cpu) {
				carryVal := 0
				if c.getCFlag() {
					carryVal = 0x80
				}

				orgVal := c.readFromBus(c.HL.getAll())
				val := orgVal>>1 | uint8(carryVal)

				c.setCFlag(orgVal&1 == 1)
				c.setNFlag(false)
				c.setZFlag(val == 0)
				c.setHFlag(false)

				c.writeToBus(c.HL.getAll(), val)

				c.PC += 2
				c.waitCycles += 16
			},
			toString: "RR (HL)",
		}, nil
	case 0x1F: // RR A
		return OpCode{
			execution: func(c *Cpu) {
				rr(c, c.AF.upper)
			},
			toString: "RR A",
		}, nil
	case 0x20: // SLA B
		return OpCode{
			execution: func(c *Cpu) {
				sla(c, c.BC.upper)
			},
			toString: "SLA B",
		}, nil
	case 0x21: // SLA C
		return OpCode{
			execution: func(c *Cpu) {
				sla(c, c.BC.lower)
			},
			toString: "SLA C",
		}, nil
	case 0x22: // SLA D
		return OpCode{
			execution: func(c *Cpu) {
				sla(c, c.DE.upper)
			},
			toString: "SLA D",
		}, nil
	case 0x23: // SLA E
		return OpCode{
			execution: func(c *Cpu) {
				sla(c, c.DE.lower)
			},
			toString: "SLA E",
		}, nil
	case 0x24: // SLA H
		return OpCode{
			execution: func(c *Cpu) {
				sla(c, c.HL.upper)
			},
			toString: "SLA H",
		}, nil
	case 0x25: // SLA L
		return OpCode{
			execution: func(c *Cpu) {
				sla(c, c.HL.lower)
			},
			toString: "SLA L",
		}, nil
	case 0x26: // SLA (HL)
		return OpCode{
			execution: func(c *Cpu) {
				val := c.readFromBus(c.HL.getAll())

				c.setCFlag(val&0x80 == 0x80)
				val <<= 1
				c.setZFlag(val == 0)
				c.setNFlag(false)
				c.setHFlag(false)

				c.writeToBus(c.HL.getAll(), val)

				c.PC += 2
				c.waitCycles += 16

			},
			toString: "SLA (HL)",
		}, nil
	case 0x27: // SLA A
		return OpCode{
			execution: func(c *Cpu) {
				sla(c, c.AF.upper)
			},
			toString: "SLA A",
		}, nil
	case 0x28: // SRA B
		return OpCode{
			execution: func(c *Cpu) {
				sra(c, c.BC.upper)
			},
			toString: "SRA B",
		}, nil
	case 0x29: // SRA C
		return OpCode{
			execution: func(c *Cpu) {
				sra(c, c.BC.lower)
			},
			toString: "SRA C",
		}, nil
	case 0x2A: // SRA D
		return OpCode{
			execution: func(c *Cpu) {
				sra(c, c.DE.upper)
			},
			toString: "SRA D",
		}, nil
	case 0x2B: // SRA E
		return OpCode{
			execution: func(c *Cpu) {
				sra(c, c.DE.lower)
			},
			toString: "SRA E",
		}, nil
	case 0x2C: // SRA H
		return OpCode{
			execution: func(c *Cpu) {
				sra(c, c.HL.upper)
			},
			toString: "SRA H",
		}, nil
	case 0x2D: // SRA L
		return OpCode{
			execution: func(c *Cpu) {
				sra(c, c.HL.lower)
			},
			toString: "SRA L",
		}, nil
	case 0x2E: // SRA (HL)
		return OpCode{
			execution: func(c *Cpu) {
				val := c.readFromBus(c.HL.getAll())
				c.setCFlag(val&1 == 1)

				if val&0x80 == 0x80 {
					val >>= 1
					val |= 0x80
				} else {
					val >>= 1
				}

				c.setZFlag(val == 0)
				c.setNFlag(false)
				c.setHFlag(false)

				c.writeToBus(c.HL.getAll(), val)

				c.PC += 2
				c.waitCycles += 16
			},
			toString: "SRA (HL)",
		}, nil
	case 0x2F: // SRA A
		return OpCode{
			execution: func(c *Cpu) {
				sra(c, c.AF.upper)
			},
			toString: "SRA A",
		}, nil
	case 0x30: // SWAP B
		return OpCode{
			execution: func(c *Cpu) {
				swap(c, c.BC.upper)
			},
			toString: "SWAP B",
		}, nil
	case 0x31: // SWAP C
		return OpCode{
			execution: func(c *Cpu) {
				swap(c, c.BC.lower)
			},
			toString: "SWAP C",
		}, nil
	case 0x32: // SWAP D
		return OpCode{
			execution: func(c *Cpu) {
				swap(c, c.DE.upper)
			},
			toString: "SWAP D",
		}, nil
	case 0x33: // SWAP E
		return OpCode{
			execution: func(c *Cpu) {
				swap(c, c.DE.lower)
			},
			toString: "SWAP E",
		}, nil
	case 0x34: // SWAP H
		return OpCode{
			execution: func(c *Cpu) {
				swap(c, c.HL.upper)
			},
			toString: "SWAP H",
		}, nil
	case 0x35: // SWAP L
		return OpCode{
			execution: func(c *Cpu) {
				swap(c, c.HL.lower)
			},
			toString: "SWAP L",
		}, nil
	case 0x36: // SWAP (HL)
		return OpCode{
			execution: func(c *Cpu) {
				val := c.readFromBus(c.HL.getAll())

				low := val & 0x0F
				high := (val & 0xF0) >> 4
				val = low<<4 | high

				c.setCFlag(false)
				c.setZFlag(val == 0)
				c.setNFlag(false)
				c.setHFlag(false)

				c.writeToBus(c.HL.getAll(), val)

				c.PC += 2
				c.waitCycles += 16
			},
			toString: "SWAP (HL)",
		}, nil
	case 0x37: // SWAP A
		return OpCode{
			execution: func(c *Cpu) {
				swap(c, c.AF.upper)
			},
			toString: "SWAP A",
		}, nil
	case 0x38: // SRL B
		return OpCode{
			execution: func(c *Cpu) {
				srl(c, c.BC.upper)
			},
			toString: "SRL B",
		}, nil
	case 0x39: // SRL C
		return OpCode{
			execution: func(c *Cpu) {
				srl(c, c.BC.lower)
			},
			toString: "SRL C",
		}, nil
	case 0x3A: // SRL D
		return OpCode{
			execution: func(c *Cpu) {
				srl(c, c.DE.upper)
			},
			toString: "SRL D",
		}, nil
	case 0x3B: // SRL E
		return OpCode{
			execution: func(c *Cpu) {
				srl(c, c.DE.lower)
			},
			toString: "SRL E",
		}, nil
	case 0x3C: // SRL H
		return OpCode{
			execution: func(c *Cpu) {
				srl(c, c.HL.upper)
			},
			toString: "SRL H",
		}, nil
	case 0x3D: // SRL L
		return OpCode{
			execution: func(c *Cpu) {
				srl(c, c.HL.lower)
			},
			toString: "SRL L",
		}, nil
	case 0x3E: // SRL (HL)
		return OpCode{
			execution: func(c *Cpu) {
				val := c.readFromBus(c.HL.getAll())

				c.setCFlag(val&1 == 1)
				val >>= 1
				c.setZFlag(val == 0)
				c.setNFlag(false)
				c.setHFlag(false)

				c.writeToBus(c.HL.getAll(), val)

				c.PC += 2
				c.waitCycles += 16
			},
			toString: "SRL (HL)",
		}, nil
	case 0x3F: // SRL A
		return OpCode{
			execution: func(c *Cpu) {
				srl(c, c.AF.upper)
			},
			toString: "SRL A",
		}, nil
	case 0x40: // BIT 0,B
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 0, c.BC.upper)
			},
			toString: "BIT 0,B",
		}, nil
	case 0x41: // BIT 0,C
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 0, c.BC.lower)
			},
			toString: "BIT 0,C",
		}, nil
	case 0x42: // BIT 0,D
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 0, c.DE.upper)
			},
			toString: "BIT 0,D",
		}, nil
	case 0x43: // BIT 0,E
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 0, c.DE.lower)
			},
			toString: "BIT 0,E",
		}, nil
	case 0x44: // BIT 0,H
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 0, c.HL.upper)
			},
			toString: "BIT 0,H",
		}, nil
	case 0x45: // BIT 0,L
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 0, c.HL.lower)
			},
			toString: "BIT 0,L",
		}, nil
	case 0x46: // BIT 0,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				bitHL(c, 0)
			},
			toString: "BIT 0,(HL)",
		}, nil
	case 0x47: // BIT 0,A
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 0, c.AF.upper)
			},
			toString: "BIT 0,A",
		}, nil
	case 0x48: // BIT 1,B
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 1, c.BC.upper)
			},
			toString: "BIT 1,B",
		}, nil
	case 0x49: // BIT 1,C
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 1, c.BC.lower)
			},
			toString: "BIT 1,C",
		}, nil
	case 0x4A: // BIT 1,D
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 1, c.DE.upper)
			},
			toString: "BIT 1,D",
		}, nil
	case 0x4B: // BIT 1,E
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 1, c.DE.lower)
			},
			toString: "BIT 1,E",
		}, nil
	case 0x4C: // BIT 1,H
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 1, c.HL.upper)
			},
			toString: "BIT 1,H",
		}, nil
	case 0x4D: // BIT 1,L
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 1, c.HL.lower)
			},
			toString: "BIT 1,L",
		}, nil
	case 0x4E: // BIT 1,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				bitHL(c, 1)
			},
			toString: "BIT 1,(HL)",
		}, nil
	case 0x4F: // BIT 1,A
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 1, c.AF.upper)
			},
			toString: "BIT 1,A",
		}, nil
	case 0x50: // BIT 2,B
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 2, c.BC.upper)
			},
			toString: "BIT 2,B",
		}, nil
	case 0x51: // BIT 2,C
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 2, c.BC.lower)
			},
			toString: "BIT 2,C",
		}, nil
	case 0x52: // BIT 2,D
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 2, c.DE.upper)
			},
			toString: "BIT 2,D",
		}, nil
	case 0x53: // BIT 2,E
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 2, c.DE.lower)
			},
			toString: "BIT 2,E",
		}, nil
	case 0x54: // BIT 2,H
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 2, c.HL.upper)
			},
			toString: "BIT 2,H",
		}, nil
	case 0x55: // BIT 2,L
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 2, c.HL.lower)
			},
			toString: "BIT 2,L",
		}, nil
	case 0x56: // BIT 2,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				bitHL(c, 2)
			},
			toString: "BIT 2,(HL)",
		}, nil
	case 0x57: // BIT 2,A
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 2, c.AF.upper)
			},
			toString: "BIT 2,A",
		}, nil
	case 0x58: // BIT 3,B
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 3, c.BC.upper)
			},
			toString: "BIT 3,B",
		}, nil
	case 0x59: // BIT 3,C
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 3, c.BC.lower)
			},
			toString: "BIT 3,C",
		}, nil
	case 0x5A: // BIT 3,D
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 3, c.DE.upper)
			},
			toString: "BIT 3,D",
		}, nil
	case 0x5B: // BIT 3,E
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 3, c.DE.lower)
			},
			toString: "BIT 3,E",
		}, nil
	case 0x5C: // BIT 3,H
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 3, c.HL.upper)
			},
			toString: "BIT 3,H",
		}, nil
	case 0x5D: // BIT 3,L
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 3, c.HL.lower)
			},
			toString: "BIT 3,L",
		}, nil
	case 0x5E: // BIT 3,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				bitHL(c, 3)
			},
			toString: "BIT 3,(HL)",
		}, nil
	case 0x5F: // BIT 3,A
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 3, c.AF.upper)
			},
			toString: "BIT 3,A",
		}, nil
	case 0x60: // BIT 4,B
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 4, c.BC.upper)
			},
			toString: "BIT 4,B",
		}, nil
	case 0x61: // BIT 4,C
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 4, c.BC.lower)
			},
			toString: "BIT 4,C",
		}, nil
	case 0x62: // BIT 4,D
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 4, c.DE.upper)
			},
			toString: "BIT 4,D",
		}, nil
	case 0x63: // BIT 4,E
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 4, c.DE.lower)
			},
			toString: "BIT 4,E",
		}, nil
	case 0x64: // BIT 4,H
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 4, c.HL.upper)
			},
			toString: "BIT 4,H",
		}, nil
	case 0x65: // BIT 4,L
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 4, c.HL.lower)
			},
			toString: "BIT 4,L",
		}, nil
	case 0x66: // BIT 4,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				bitHL(c, 4)
			},
			toString: "BIT 4,(HL)",
		}, nil
	case 0x67: // BIT 4,A
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 4, c.AF.upper)
			},
			toString: "BIT 4,A",
		}, nil
	case 0x68: // BIT 5,B
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 5, c.BC.upper)
			},
			toString: "BIT 5,B",
		}, nil
	case 0x69: // BIT 5,C
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 5, c.BC.lower)
			},
			toString: "BIT 5,C",
		}, nil
	case 0x6A: // BIT 5,D
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 5, c.DE.upper)
			},
			toString: "BIT 5,D",
		}, nil
	case 0x6B: // BIT 5,E
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 5, c.DE.lower)
			},
			toString: "BIT 5,E",
		}, nil
	case 0x6C: // BIT 5,H
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 5, c.HL.upper)
			},
			toString: "BIT 5,H",
		}, nil
	case 0x6D: // BIT 5,L
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 5, c.HL.lower)
			},
			toString: "BIT 5,L",
		}, nil
	case 0x6E: // BIT 5,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				bitHL(c, 5)
			},
			toString: "BIT 5,(HL)",
		}, nil
	case 0x6F: // BIT 5,A
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 5, c.AF.upper)
			},
			toString: "BIT 5,A",
		}, nil
	case 0x70: // BIT 6,B
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 6, c.BC.upper)
			},
			toString: "BIT 6,B",
		}, nil
	case 0x71: // BIT 6,C
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 6, c.BC.lower)
			},
			toString: "BIT 6,C",
		}, nil
	case 0x72: // BIT 6,D
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 6, c.DE.upper)
			},
			toString: "BIT 6,D",
		}, nil
	case 0x73: // BIT 6,E
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 6, c.DE.lower)
			},
			toString: "BIT 6,E",
		}, nil
	case 0x74: // BIT 6,H
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 6, c.HL.upper)
			},
			toString: "BIT 6,H",
		}, nil
	case 0x75: // BIT 6,L
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 6, c.HL.lower)
			},
			toString: "BIT 6,L",
		}, nil
	case 0x76: // BIT 6,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				bitHL(c, 6)
			},
			toString: "BIT 6,(HL)",
		}, nil
	case 0x77: // BIT 6,A
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 6, c.AF.upper)
			},
			toString: "BIT 6,A",
		}, nil
	case 0x78: // BIT 7,B
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 7, c.BC.upper)
			},
			toString: "BIT 7,B",
		}, nil
	case 0x79: // BIT 7,C
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 7, c.BC.lower)
			},
			toString: "BIT 7,C",
		}, nil
	case 0x7A: // BIT 7,D
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 7, c.DE.upper)
			},
			toString: "BIT 7,D",
		}, nil
	case 0x7B: // BIT 7,E
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 7, c.DE.lower)
			},
			toString: "BIT 7,E",
		}, nil
	case 0x7C: // BIT 7,H
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 7, c.HL.upper)
			},
			toString: "BIT 7,H",
		}, nil
	case 0x7D: // BIT 7,L
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 7, c.HL.lower)
			},
			toString: "BIT 7,L",
		}, nil
	case 0x7E: // BIT 7,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				bitHL(c, 7)
			},
			toString: "BIT 7,(HL)",
		}, nil
	case 0x7F: // BIT 7,A
		return OpCode{
			execution: func(c *Cpu) {
				bit(c, 7, c.AF.upper)
			},
			toString: "BIT 7,A",
		}, nil
	case 0x80: // RES 0,B
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 0, c.BC.upper)
			},
			toString: "RES 0,B",
		}, nil
	case 0x81: // RES 0,C
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 0, c.BC.lower)
			},
			toString: "RES 0,C",
		}, nil
	case 0x82: // RES 0,D
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 0, c.DE.upper)
			},
			toString: "RES 0,D",
		}, nil
	case 0x83: // RES 0,E
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 0, c.DE.lower)
			},
			toString: "RES 0,E",
		}, nil
	case 0x84: // RES 0,H
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 0, c.HL.upper)
			},
			toString: "RES 0,H",
		}, nil
	case 0x85: // RES 0,L
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 0, c.HL.lower)
			},
			toString: "RES 0,L",
		}, nil
	case 0x86: // RES 0,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				resHL(c, 0)
			},
			toString: "RES 0,(HL)",
		}, nil
	case 0x87: // RES 0,A
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 0, c.AF.upper)
			},
			toString: "RES 0,A",
		}, nil
	case 0x88: // RES 1,B
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 1, c.BC.upper)
			},
			toString: "RES 1,B",
		}, nil
	case 0x89: // RES 1,C
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 1, c.BC.lower)
			},
			toString: "RES 1,C",
		}, nil
	case 0x8A: // RES 1,D
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 1, c.DE.upper)
			},
			toString: "RES 1,D",
		}, nil
	case 0x8B: // RES 1,E
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 1, c.DE.lower)
			},
			toString: "RES 1,E",
		}, nil
	case 0x8C: // RES 1,H
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 1, c.HL.upper)
			},
			toString: "RES 1,H",
		}, nil
	case 0x8D: // RES 1,L
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 1, c.HL.lower)
			},
			toString: "RES 1,L",
		}, nil
	case 0x8E: // RES 1,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				resHL(c, 1)
			},
			toString: "RES 1,(HL)",
		}, nil
	case 0x8F: // RES 1,A
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 1, c.AF.upper)
			},
			toString: "RES 1,A",
		}, nil

	case 0x90: // RES 2,B
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 2, c.BC.upper)
			},
			toString: "RES 2,B",
		}, nil
	case 0x91: // RES 2,C
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 2, c.BC.lower)
			},
			toString: "RES 2,C",
		}, nil
	case 0x92: // RES 2,D
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 2, c.DE.upper)
			},
			toString: "RES 2,D",
		}, nil
	case 0x93: // RES 2,E
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 2, c.DE.lower)
			},
			toString: "RES 2,E",
		}, nil
	case 0x94: // RES 2,H
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 2, c.HL.upper)
			},
			toString: "RES 2,H",
		}, nil
	case 0x95: // RES 2,L
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 2, c.HL.lower)
			},
			toString: "RES 2,L",
		}, nil
	case 0x96: // RES 2,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				resHL(c, 2)
			},
			toString: "RES 2,(HL)",
		}, nil
	case 0x97: // RES 2,A
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 2, c.AF.upper)
			},
			toString: "RES 2,a",
		}, nil
	case 0x98: // RES 3,B
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 3, c.BC.upper)
			},
			toString: "RES 3,B",
		}, nil
	case 0x99: // RES 3,C
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 3, c.BC.lower)
			},
			toString: "RES 3,C",
		}, nil
	case 0x9A: // RES 3,D
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 3, c.DE.upper)
			},
			toString: "RES 3,D",
		}, nil
	case 0x9B: // RES 3,E
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 3, c.DE.lower)
			},
			toString: "RES 3,E",
		}, nil
	case 0x9C: // RES 3,H
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 3, c.HL.upper)
			},
			toString: "RES 3,H",
		}, nil
	case 0x9D: // RES 3,L
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 3, c.HL.lower)
			},
			toString: "RES 3,L",
		}, nil
	case 0x9E: // RES 3,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				resHL(c, 3)
			},
			toString: "RES 3,(HL)",
		}, nil
	case 0x9F: // RES 3,A
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 3, c.AF.upper)
			},
			toString: "RES 3,A",
		}, nil
	case 0xA0: // RES 4,B
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 4, c.BC.upper)
			},
			toString: "RES 4,B",
		}, nil
	case 0xA1: // RES 4,C
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 4, c.BC.lower)
			},
			toString: "RES 4,C",
		}, nil
	case 0xA2: // RES 4,D
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 4, c.DE.upper)
			},
			toString: "RES 4,D",
		}, nil
	case 0xA3: // RES 4,E
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 4, c.DE.lower)
			},
			toString: "RES 4,E",
		}, nil
	case 0xA4: // RES 4,H
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 4, c.HL.upper)
			},
			toString: "RES 4,H",
		}, nil
	case 0xA5: // RES 4,L
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 4, c.HL.lower)
			},
			toString: "RES 4,L",
		}, nil
	case 0xA6: // RES 4,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				resHL(c, 4)
			},
			toString: "RES 4,(HL)",
		}, nil
	case 0xA7: // RES 4,A
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 4, c.AF.upper)
			},
			toString: "RES 4,A",
		}, nil
	case 0xA8: // RES 5,B
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 5, c.BC.upper)
			},
			toString: "RES 5,B",
		}, nil
	case 0xA9: // RES 5,C
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 5, c.BC.lower)
			},
			toString: "RES 5,C",
		}, nil
	case 0xAA: // RES 5,D
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 5, c.DE.upper)
			},
			toString: "RES 5,D",
		}, nil
	case 0xAB: // RES 5,E
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 5, c.DE.lower)
			},
			toString: "RES 5,E",
		}, nil
	case 0xAC: // RES 5,H
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 5, c.HL.upper)
			},
			toString: "RES 5,H",
		}, nil
	case 0xAD: // RES 5,L
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 5, c.HL.lower)
			},
			toString: "RES 5,L",
		}, nil
	case 0xAE: // RES 5,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				resHL(c, 5)
			},
			toString: "RES 5,(HL)",
		}, nil
	case 0xAF: // RES 5,A
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 5, c.AF.upper)
			},
			toString: "RES 5,A",
		}, nil
	case 0xB0: // RES 6,B
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 6, c.BC.upper)
			},
			toString: "RES 6,B",
		}, nil
	case 0xB1: // RES 6,C
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 6, c.BC.lower)
			},
			toString: "RES 6,C",
		}, nil
	case 0xB2: // RES 6,D
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 6, c.DE.upper)
			},
			toString: "RES 6,D",
		}, nil
	case 0xB3: // RES 6,E
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 6, c.DE.lower)
			},
			toString: "RES 6,E",
		}, nil
	case 0xB4: // RES 6,H
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 6, c.HL.upper)
			},
			toString: "RES 6,H",
		}, nil
	case 0xB5: // RES 6,L
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 6, c.HL.lower)
			},
			toString: "RES 6,L",
		}, nil
	case 0xB6: // RES 6,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				resHL(c, 6)
			},
			toString: "RES 6,(HL)",
		}, nil
	case 0xB7: // RES 6,A
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 6, c.AF.upper)
			},
			toString: "RES 6,(HL)",
		}, nil
	case 0xB8: // RES 7,B
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 7, c.BC.upper)
			},
			toString: "RES 7,B",
		}, nil
	case 0xB9: // RES 7,C
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 7, c.BC.lower)
			},
			toString: "RES 7,C",
		}, nil
	case 0xBA: // RES 7,D
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 7, c.DE.upper)
			},
			toString: "RES 7,D",
		}, nil
	case 0xBB: // RES 7,E
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 7, c.DE.lower)
			},
			toString: "RES 7,E",
		}, nil
	case 0xBC: // RES 7,H
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 7, c.HL.upper)
			},
			toString: "RES 7,H",
		}, nil
	case 0xBD: // RES 7,L
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 7, c.HL.lower)
			},
			toString: "RES 7,L",
		}, nil
	case 0xBE: // RES 7,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				resHL(c, 7)
			},
			toString: "RES 7,(HL)",
		}, nil
	case 0xBF: // RES 7,A
		return OpCode{
			execution: func(c *Cpu) {
				res(c, 7, c.AF.upper)
			},
			toString: "RES 7,A",
		}, nil
	case 0xC0: // SET 0,B
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 0, c.BC.upper)
			},
			toString: "SET 0,B",
		}, nil
	case 0xC1: // SET 0,C
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 0, c.BC.lower)
			},
			toString: "SET 0,C",
		}, nil
	case 0xC2: // SET 0,D
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 0, c.DE.upper)
			},
			toString: "SET 0,D",
		}, nil
	case 0xC3: // SET 0,E
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 0, c.DE.lower)
			},
			toString: "SET 0,E",
		}, nil
	case 0xC4: // SET 0,H
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 0, c.HL.upper)
			},
			toString: "SET 0,H",
		}, nil
	case 0xC5: // SET 0,L
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 0, c.HL.lower)
			},
			toString: "SET 0,L",
		}, nil
	case 0xC6: // SET 0,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				setHL(c, 0)
			},
			toString: "SET 0,(HL)",
		}, nil
	case 0xC7: // SET 0,A
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 0, c.AF.upper)
			},
			toString: "SET 0,A",
		}, nil
	case 0xC8: // SET 1,B
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 1, c.BC.upper)
			},
			toString: "SET 1,B",
		}, nil
	case 0xC9: // SET 1,C
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 1, c.BC.lower)
			},
			toString: "SET 1,C",
		}, nil
	case 0xCA: // SET 1,D
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 1, c.DE.upper)
			},
			toString: "SET 1,D",
		}, nil
	case 0xCB: // SET 1,E
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 1, c.DE.lower)
			},
			toString: "SET 1,E",
		}, nil
	case 0xCC: // SET 1,H
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 1, c.HL.upper)
			},
			toString: "SET 1,H",
		}, nil
	case 0xCD: // SET 1,L
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 1, c.HL.lower)
			},
			toString: "SET 1,L",
		}, nil
	case 0xCE: // SET 1,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				setHL(c, 1)
			},
			toString: "SET 1,(HL)",
		}, nil
	case 0xCF: // SET 1,A
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 1, c.AF.upper)
			},
			toString: "SET 1,A",
		}, nil
	case 0xD0: // SET 2,B
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 2, c.BC.upper)
			},
			toString: "SET 2,B",
		}, nil
	case 0xD1: // SET 2,C
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 2, c.BC.lower)
			},
			toString: "SET 2,C",
		}, nil
	case 0xD2: // SET 2,D
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 2, c.DE.upper)
			},
			toString: "SET 2,D",
		}, nil
	case 0xD3: // SET 2,E
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 2, c.DE.lower)
			},
			toString: "SET 2,E",
		}, nil
	case 0xD4: // SET 2,H
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 2, c.HL.upper)
			},
			toString: "SET 2,H",
		}, nil
	case 0xD5: // SET 2,L
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 2, c.HL.lower)
			},
			toString: "SET 2,L",
		}, nil
	case 0xD6: // SET 2,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				setHL(c, 2)
			},
			toString: "SET 2,(HL)",
		}, nil
	case 0xD7: // SET 2,A
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 2, c.AF.upper)
			},
			toString: "SET 2,A",
		}, nil
	case 0xD8: // SET 3,B
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 3, c.BC.upper)
			},
			toString: "SET 3,B",
		}, nil
	case 0xD9: // SET 3,C
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 3, c.BC.lower)
			},
			toString: "SET 3,C",
		}, nil
	case 0xDA: // SET 3,D
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 3, c.DE.upper)
			},
			toString: "SET 3,D",
		}, nil
	case 0xDB: // SET 3,E
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 3, c.DE.lower)
			},
			toString: "SET 3,E",
		}, nil
	case 0xDC: // SET 3,H
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 3, c.HL.upper)
			},
			toString: "SET 3,H",
		}, nil
	case 0xDD: // SET 3,L
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 3, c.HL.lower)
			},
			toString: "SET 3,L",
		}, nil
	case 0xDE: // SET 3,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				setHL(c, 3)
			},
			toString: "SET 3,(HL)",
		}, nil
	case 0xDF: // SET 3,A
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 3, c.AF.upper)
			},
			toString: "SET 3,A",
		}, nil
	case 0xE0: // SET 4,B
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 4, c.BC.upper)
			},
			toString: "SET 4,B",
		}, nil
	case 0xE1: // SET 4,C
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 4, c.BC.lower)
			},
			toString: "SET 4,C",
		}, nil
	case 0xE2: // SET 4,D
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 4, c.DE.upper)
			},
			toString: "SET 4,D",
		}, nil
	case 0xE3: // SET 4,E
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 4, c.DE.lower)
			},
			toString: "SET 4,E",
		}, nil
	case 0xE4: // SET 4,H
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 4, c.HL.upper)
			},
			toString: "SET 4,H",
		}, nil
	case 0xE5: // SET 4,L
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 4, c.HL.lower)
			},
			toString: "SET 4,L",
		}, nil
	case 0xE6: // SET 4,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				setHL(c, 4)
			},
			toString: "SET 4,(HL)",
		}, nil
	case 0xE7: // SET 4,A
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 4, c.AF.upper)
			},
			toString: "SET 4,A",
		}, nil
	case 0xE8: // SET 5,B
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 5, c.BC.upper)
			},
			toString: "SET 5,B",
		}, nil
	case 0xE9: // SET 5,C
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 5, c.BC.lower)
			},
			toString: "SET 5,C",
		}, nil
	case 0xEA: // SET 5,D
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 5, c.DE.upper)
			},
			toString: "SET 5,D",
		}, nil
	case 0xEB: // SET 5,E
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 5, c.DE.lower)
			},
			toString: "SET 5,E",
		}, nil
	case 0xEC: // SET 5,H
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 5, c.HL.upper)
			},
			toString: "SET 5,H",
		}, nil
	case 0xED: // SET 5,L
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 5, c.HL.lower)
			},
			toString: "SET 5,L",
		}, nil
	case 0xEE: // SET 5,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				setHL(c, 5)
			},
			toString: "SET 5,(HL)",
		}, nil
	case 0xEF: // SET 5,A
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 5, c.AF.upper)
			},
			toString: "SET 5,A",
		}, nil
	case 0xF0: // SET 6,B
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 6, c.BC.upper)
			},
			toString: "SET 6,B",
		}, nil
	case 0xF1: // SET 6,C
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 6, c.BC.lower)
			},
			toString: "SET 6,C",
		}, nil
	case 0xF2: // SET 6,D
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 6, c.DE.upper)
			},
			toString: "SET 6,D",
		}, nil
	case 0xF3: // SET 6,E
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 6, c.DE.lower)
			},
			toString: "SET 6,E",
		}, nil
	case 0xF4: // SET 6,H
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 6, c.HL.upper)
			},
			toString: "SET 6,H",
		}, nil
	case 0xF5: // SET 6,L
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 6, c.HL.lower)
			},
			toString: "SET 6,L",
		}, nil
	case 0xF6: // SET 6,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				setHL(c, 6)
			},
			toString: "SET 6,(HL)",
		}, nil
	case 0xF7: // SET 6,A
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 6, c.AF.upper)
			},
			toString: "SET 6,A",
		}, nil
	case 0xF8: // SET 7,B
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 7, c.BC.upper)
			},
			toString: "SET 7,B",
		}, nil
	case 0xF9: // SET 7,C
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 7, c.BC.lower)
			},
			toString: "SET 7,C",
		}, nil
	case 0xFA: // SET 7,D
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 7, c.DE.upper)
			},
			toString: "SET 7,D",
		}, nil
	case 0xFB: // SET 7,E
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 7, c.DE.lower)
			},
			toString: "SET 7,E",
		}, nil
	case 0xFC: // SET 7,H
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 7, c.HL.upper)
			},
			toString: "SET 7,H",
		}, nil
	case 0xFD: // SET 7,L
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 7, c.HL.lower)
			},
			toString: "SET 7,L",
		}, nil
	case 0xFE: // SET 7,(HL)
		return OpCode{
			execution: func(c *Cpu) {
				setHL(c, 7)
			},
			toString: "SET 7,(HL)",
		}, nil
	case 0xFF: // SET 7,A
		return OpCode{
			execution: func(c *Cpu) {
				set(c, 7, c.AF.upper)
			},
			toString: "SET 7,A",
		}, nil
	default:
		return OpCode{}, fmt.Errorf("opcode %d not recognized", code)
	}
}

func setHL(c *Cpu, bitNum byte) {
	val := c.readFromBus(c.HL.getAll())
	val |= 1 << bitNum
	c.writeToBus(c.HL.getAll(), val)

	c.PC += 2
	c.waitCycles += 16
}

func set(c *Cpu, bitNum byte, halfRegister *HalfRegister) {
	halfRegister.value |= 1 << bitNum

	c.PC += 2
	c.waitCycles += 8
}

func resHL(c *Cpu, bitNum byte) {
	val := c.readFromBus(c.HL.getAll())
	val &= ^(1 << bitNum)
	c.writeToBus(c.HL.getAll(), val)

	c.PC += 2
	c.waitCycles += 16
}

func res(c *Cpu, bitNum byte, halfRegister *HalfRegister) {
	halfRegister.value &= ^(1 << bitNum)

	c.PC += 2
	c.waitCycles += 8
}

func bitHL(c *Cpu, bitNum byte) {
	val := c.readFromBus(c.HL.getAll())
	c.setZFlag((val>>bitNum)&1 == 0)
	c.setNFlag(false)
	c.setHFlag(true)

	c.PC += 2
	c.waitCycles += 12
}

func bit(c *Cpu, bitNum byte, halfRegister *HalfRegister) {
	c.setZFlag((halfRegister.value>>bitNum)&1 == 0)
	c.setNFlag(false)
	c.setHFlag(true)

	c.PC += 2
	c.waitCycles += 8
}

func srl(c *Cpu, halfRegister *HalfRegister) {
	c.setCFlag(halfRegister.value&1 == 1)
	halfRegister.value >>= 1
	c.setZFlag(halfRegister.value == 0)
	c.setNFlag(false)
	c.setHFlag(false)

	c.PC += 2
	c.waitCycles += 8
}

func swap(c *Cpu, halfRegister *HalfRegister) {
	low := halfRegister.value & 0x0F
	high := (halfRegister.value & 0xF0) >> 4
	halfRegister.value = low<<4 | high

	c.setCFlag(false)
	c.setZFlag(halfRegister.value == 0)
	c.setNFlag(false)
	c.setHFlag(false)

	c.PC += 2
	c.waitCycles += 8
}

func sra(c *Cpu, halfRegister *HalfRegister) {
	c.setCFlag(halfRegister.value&1 == 1)

	if halfRegister.value&0x80 == 0x80 {
		halfRegister.value >>= 1
		halfRegister.value |= 0x80
	} else {
		halfRegister.value >>= 1
	}

	c.setZFlag(halfRegister.value == 0)
	c.setNFlag(false)
	c.setHFlag(false)

	c.PC += 2
	c.waitCycles += 8
}

func sla(c *Cpu, halfRegister *HalfRegister) {
	c.setCFlag(halfRegister.value&0x80 == 0x80)
	halfRegister.value <<= 1
	c.setZFlag(halfRegister.value == 0)
	c.setNFlag(false)
	c.setHFlag(false)

	c.PC += 2
	c.waitCycles += 8
}

func rr(c *Cpu, halfRegister *HalfRegister) {
	carryVal := 0
	if c.getCFlag() {
		carryVal = 0x80
	}

	orgVal := halfRegister.value
	val := halfRegister.value>>1 | uint8(carryVal)

	c.setCFlag(orgVal&1 == 1)
	c.setNFlag(false)
	c.setZFlag(val == 0)
	c.setHFlag(false)

	halfRegister.value = val

	c.PC += 2
	c.waitCycles += 8
}

func rl(c *Cpu, halfRegister *HalfRegister) {
	carryVal := 0
	if c.getCFlag() {
		carryVal = 1
	}
	val := halfRegister.value<<1 | uint8(carryVal)

	c.setCFlag(halfRegister.value&0x80 == 0x80)
	c.setNFlag(false)
	c.setZFlag(val == 0)
	c.setHFlag(false)

	halfRegister.value = val

	c.PC += 2
	c.waitCycles += 8
}

func rrc(c *Cpu, halfRegister *HalfRegister) {
	orgVal := halfRegister.value
	halfRegister.value >>= 1

	if orgVal&1 == 1 {
		halfRegister.value |= 0x80
		c.setCFlag(true)
	} else {
		c.setCFlag(false)
	}

	c.setNFlag(false)
	c.setZFlag(halfRegister.value == 0)
	c.setHFlag(false)

	c.PC += 2
	c.waitCycles += 8
}

func rlc(c *Cpu, halfRegister *HalfRegister) {
	orgVal := halfRegister.value
	halfRegister.value <<= 1

	if orgVal&0x80 == 0x80 {
		halfRegister.value |= 1
		c.setCFlag(true)
	} else {
		c.setCFlag(false)
	}

	c.setNFlag(false)
	c.setZFlag(halfRegister.value == 0)
	c.setHFlag(false)

	c.PC += 2
	c.waitCycles += 8
}
