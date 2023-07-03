package ppu

func getColour(colourIdx byte) uint32 {
	switch colourIdx {
	case 0: // white
		return 0xFFFFFFFF
	case 1: // light grey
		return 0xD3D3D3FF
	case 2: // dark grey
		return 0x808080FF
	case 3: // black
		return 0x000000FF
	default:
		return 0xFF0000FF
	}
}
