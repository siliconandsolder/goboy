package cartridge

const (
	// read only
	LOWER_ROM_BANK_END   = 0x3FFF
	UPPER_ROM_BANK_START = 0x4000
	UPPER_ROM_BANK_END   = 0x7FFF

	RAM_BANK_START = 0xA000
	RAM_BANK_END   = 0xBFFF

	// registers
	RAM_ENABLE_END = 0x1FFF

	ROM_BANK_SELECT_START = 0x2000
	ROM_BANK_SELECT_END   = 0x3FFF

	RAM_BANK_SELECT_START = 0x4000
	RAM_BANK_SELECT_END   = 0x5FFF

	UPPER_BANK_SELECT_START = 0x6000
	UPPER_BANK_SELECT_END   = 0x7FFF
)

// MBC returns real address for ROM or RAM
type MBC interface {
	Read(addr uint16) (uint32, bool)
	Write(addr uint16, data byte) (uint32, bool)
}
