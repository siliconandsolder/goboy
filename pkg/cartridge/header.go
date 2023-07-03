package cartridge

const (
	LOGO            = 0x04
	TITLE           = 0x34
	MAN_CODE        = 0x3F
	CGB_FLAG        = 0x43
	NEW_LIC_CODE    = 0x44
	SGB_CODE        = 0x46
	CART_TYPE       = 0x47
	ROM_SIZE        = 0x48
	RAM_SIZE        = 0x49
	DEST_CODE       = 0x4A
	OLD_LIC_CODE    = 0x4B
	ROM_VERSION     = 0x4C
	HEADER_CHECKSUM = 0x4D

	CGB_ONLY_CODE = 0xC0
)

var LogoBytes = []byte{
	0xCE, 0xED, 0x66, 0x66, 0xCC, 0x0D, 0x00, 0x0B, 0x03, 0x73, 0x00, 0x83, 0x00, 0x0C, 0x00, 0x0D,
	0x00, 0x08, 0x11, 0x1F, 0x88, 0x89, 0x00, 0x0E, 0xDC, 0xCC, 0x6E, 0xE6, 0xDD, 0xDD, 0xD9, 0x99,
	0xBB, 0xBB, 0x67, 0x63, 0x6E, 0x0E, 0xEC, 0xCC, 0xDD, 0xDC, 0x99, 0x9F, 0xBB, 0xB9, 0x33, 0x3E,
}

type Header struct {
	Title          string
	ManCode        string
	CGBFlag        bool
	LicenceCode    byte
	SGBFlag        bool
	CartType       byte
	RomSize        RomInfo
	RamSize        RamInfo
	DestCode       byte
	OldLicenceCode byte
	RomVersion     byte
	HeaderChecksum byte
	GlobalChecksum byte
}

type RomInfo struct {
	Size     uint32
	NumBanks uint16
}

type RamInfo struct {
	Size     uint32
	NumBanks uint16
}

func NewHeader(data []byte) *Header {
	return &Header{
		Title:          string(data[TITLE:MAN_CODE]),
		ManCode:        string(data[MAN_CODE:CGB_FLAG]),
		CGBFlag:        data[CGB_FLAG] == CGB_ONLY_CODE,
		LicenceCode:    0,
		SGBFlag:        data[SGB_CODE] == 0x3,
		CartType:       data[CART_TYPE],
		RomSize:        getRomData(data[ROM_SIZE]),
		RamSize:        getRamData(data[RAM_SIZE]),
		DestCode:       data[DEST_CODE],
		OldLicenceCode: 0,
		RomVersion:     data[ROM_VERSION],
		HeaderChecksum: data[HEADER_CHECKSUM],
		GlobalChecksum: 0,
	}
}

func getRomData(romValue byte) RomInfo {
	switch romValue {
	case 0x00:
		return RomInfo{
			Size:     0x8000,
			NumBanks: 2,
		}
	case 0x01:
		return RomInfo{
			Size:     0x10000,
			NumBanks: 4,
		}
	case 0x02:
		return RomInfo{
			Size:     0x20000,
			NumBanks: 8,
		}
	case 0x03:
		return RomInfo{
			Size:     0x40000,
			NumBanks: 16,
		}
	case 0x04:
		return RomInfo{
			Size:     0x80000,
			NumBanks: 32,
		}
	case 0x05:
		return RomInfo{
			Size:     0x100000,
			NumBanks: 64,
		}
	case 0x06:
		return RomInfo{
			Size:     0x200000,
			NumBanks: 128,
		}
	case 0x07:
		return RomInfo{
			Size:     0x400000,
			NumBanks: 256,
		}
	case 0x08:
		return RomInfo{
			Size:     0x800000,
			NumBanks: 512,
		}
	default:
		return RomInfo{}
	}
}

func getRamData(ramValue byte) RamInfo {
	switch ramValue {
	case 0:
	case 1:
		return RamInfo{
			Size:     0,
			NumBanks: 0,
		}
	case 2:
		return RamInfo{
			Size:     0x2000,
			NumBanks: 1,
		}
	case 3:
		return RamInfo{
			Size:     0x8000,
			NumBanks: 4,
		}
	case 4:
		return RamInfo{
			Size:     0x20000,
			NumBanks: 16,
		}
	case 5:
		return RamInfo{
			Size:     0x10000,
			NumBanks: 8,
		}
	default:
		return RamInfo{}
	}
	return RamInfo{}
}
