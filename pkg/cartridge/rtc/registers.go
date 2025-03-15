package rtc

type registers struct {
	S  byte
	M  byte
	H  byte
	DL byte
	DH byte
}

func (r *registers) GetRegisterValue(reg byte) byte {
	switch reg {
	case 0x08:
		return r.S & 0x3F
	case 0x09:
		return r.M & 0x3F
	case 0x0A:
		return r.H & 0x1F
	case 0x0B:
		return r.DL
	case 0x0C:
		return r.DH & 0xC1
	default:
		panic("unrecognized value: " + string(reg))
	}
}
