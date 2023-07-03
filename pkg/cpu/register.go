package cpu

type Register struct {
	upper *HalfRegister
	lower *HalfRegister
}

func NewRegister() *Register {
	return &Register{
		upper: NewHalfRegister(),
		lower: NewHalfRegister(),
	}
}

//func (reg *Register) getHigh() byte {
//	return byte(reg.value >> 8)
//}
//
//func (reg *Register) getLow() byte {
//	return byte(reg.value & 0x00FF)
//}

func (reg *Register) getAll() uint16 {
	return uint16(reg.upper.value)<<8 | uint16(reg.lower.value)
}

//func (reg *Register) setHigh(val byte) {
//	reg.value = (uint16(val) << 8) | (reg.value & 0x00FF)
//}
//
//func (reg *Register) setLow(val byte) {
//	reg.value = (reg.value & 0xFF00) | uint16(val)
//}

func (reg *Register) setAll(val uint16) {
	//reg.value = val
	reg.upper.value = byte(val >> 8)
	reg.lower.value = byte(val & 0x00FF)
}
