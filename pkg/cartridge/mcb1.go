package cartridge

const (
	SimpleBankMode = iota
	RamBankMode
)

type MBC1 struct {
	ramEnabled      bool
	lowerRomBankNum byte
	upperRomBankNum byte
	lastRomBank     byte
	hasRam          bool
	ramBankNum      byte
	is1MBRom        bool
	is32kRam        bool
	bankSelectMode  byte
}

func NewMBC1(romInfo RomInfo, ramInfo RamInfo) *MBC1 {
	return &MBC1{
		ramEnabled:      false,
		lowerRomBankNum: 0,
		upperRomBankNum: 0,
		lastRomBank:     byte(romInfo.NumBanks - 1),
		hasRam:          ramInfo.Size > 0,
		ramBankNum:      0,
		is1MBRom:        romInfo.Size >= 0x100000,
		is32kRam:        ramInfo.Size >= 0x8000,
		bankSelectMode:  0,
	}
}

func (mbc1 *MBC1) Read(addr uint16) (uint32, bool) {
	if addr <= LOWER_ROM_BANK_END {
		return uint32(addr) + uint32(mbc1.getLowerRomBank())*0x4000, true
	} else if addr >= UPPER_ROM_BANK_START && addr <= UPPER_ROM_BANK_END {
		return uint32(addr-0x4000) + uint32(mbc1.getUpperRomBank())*0x4000, true
	} else if mbc1.hasRam && mbc1.ramEnabled && addr >= RAM_BANK_START && addr <= RAM_BANK_END {
		return uint32(addr) + uint32(mbc1.getRamBank())*0x2000, true
	}

	return 0, true
}
func (mbc1 *MBC1) Write(addr uint16, data byte) (uint32, bool) {
	if addr <= RAM_ENABLE_END {
		if data&0x0A == 0x0A {
			mbc1.ramEnabled = true
		} else {
			mbc1.ramEnabled = false
		}
	} else if addr >= ROM_BANK_SELECT_START && addr <= ROM_BANK_SELECT_END {
		mbc1.lowerRomBankNum = data & 0x1F
	} else if addr >= RAM_BANK_SELECT_START && addr <= RAM_BANK_SELECT_END {
		mbc1.upperRomBankNum = data & 0x03
	} else if addr >= UPPER_BANK_SELECT_START && addr <= UPPER_BANK_SELECT_END {
		mbc1.bankSelectMode = data & 0x02
	}

	return 0, false
}

func (mbc1 *MBC1) getRamBank() byte {
	if mbc1.bankSelectMode == SimpleBankMode || !mbc1.is32kRam {
		return 0
	}

	return mbc1.upperRomBankNum
}

func (mbc1 *MBC1) getUpperRomBank() byte {
	bank := mbc1.lowerRomBankNum
	if bank == 0 {
		bank = 1
	}

	return bank & mbc1.lastRomBank
}

func (mbc1 *MBC1) getLowerRomBank() byte {
	if mbc1.bankSelectMode == SimpleBankMode || !mbc1.is1MBRom {
		return 0
	}

	bank := mbc1.upperRomBankNum << 5
	return bank & mbc1.lastRomBank
}
