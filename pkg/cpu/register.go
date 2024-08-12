package cpu

type Register struct {
	upper *HalfRegister
	lower *HalfRegister
}

func NewRegister() *Register {
	return &Register{
		upper: &HalfRegister{value: 0},
		lower: &HalfRegister{value: 0},
	}
}

func (reg *Register) getAll() uint16 {
	return uint16(reg.upper.value)<<8 | uint16(reg.lower.value)
}

func (reg *Register) setAll(val uint16) {
	//reg.value = val
	reg.upper.value = byte(val >> 8)
	reg.lower.value = byte(val & 0x00FF)
}
