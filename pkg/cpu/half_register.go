package cpu

type HalfRegister struct {
	value byte
}

func NewHalfRegister() *HalfRegister {
	return &HalfRegister{value: 0}
}
