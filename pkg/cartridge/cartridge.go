package cartridge

import (
	"fmt"
	"os"
)

const (
	ROM_ONLY = 0x00
	MBC_1    = 0x01
)

const (
	HEADER_START         = 0x100
	CHECKSUM_LOWER_BOUND = 0x134
	CHECKSUM_UPPER_BOUND = 0x14D

	RAM_START = 0xA000
	RAM_END   = 0xBFFF
)

type Cartridge struct {
	header *Header
	mbc    MBC
	rom    []byte
	ram    []byte
}

func NewCartridge(file string) *Cartridge {
	data, err := os.ReadFile(file)
	if err != nil {
		panic(err) // no point in continuing
	}

	header := NewHeader(data[HEADER_START:])

	if err := verifyChecksum(header.HeaderChecksum, data[CHECKSUM_LOWER_BOUND:CHECKSUM_UPPER_BOUND]); err != nil {
		panic(err)
	}

	mapper, err := getMBC(header)
	if err != nil {
		panic(err)
	}

	rom := make([]byte, header.RomSize.Size)
	ram := make([]byte, header.RamSize.Size)

	rom = data

	return &Cartridge{
		header: header,
		mbc:    mapper,
		rom:    rom,
		ram:    ram,
	}

}

func (c *Cartridge) Read(addr uint16) byte {
	realAddr := c.mbc.Read(addr)
	if addr >= RAM_START && addr <= RAM_END {
		return c.ram[realAddr]
	}

	return c.rom[realAddr]
}

func (c *Cartridge) Write(addr uint16, data byte) {
	realAddr := c.mbc.Write(addr, data)
	if addr >= RAM_START && addr <= RAM_END {
		c.ram[realAddr] = data
	}
}

func verifyChecksum(verifier byte, verifyBytes []byte) error {
	var checksum byte = 0
	for _, val := range verifyBytes {
		checksum = checksum - val - 1
	}

	if checksum != verifier {
		return fmt.Errorf("checksum is invalid. was %d, should be %d", checksum, verifier)
	}
	return nil
}

func getMBC(header *Header) (MBC, error) {
	switch header.CartType {
	case ROM_ONLY:
		return &RomOnly{}, nil
	case MBC_1:
		return NewMBC1(header.RomSize, header.RamSize), nil
	default:
		return nil, fmt.Errorf("cart type %d not yet implemented", header.CartType)
	}
}
