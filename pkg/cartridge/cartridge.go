package cartridge

import (
	"fmt"
	"github.com/siliconandsolder/go-boy/pkg/cartridge/rtc"
	"os"
)

const (
	ROM_ONLY    = 0x00
	MBC_1       = 0x01
	MBC_3_START = 0x0F
	MBC_3_END   = 0x13
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
	state  *rtc.State
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

	mapper, state, err := getMBC(header)
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
		state:  state,
	}
}

func (c *Cartridge) UpdateCounter(cycles byte) {
	if c.state != nil {
		c.state.AddCycles(cycles)
	}
}

func (c *Cartridge) Read(addr uint16) byte {
	val, isAddr := c.mbc.Read(addr)
	if isAddr {
		if addr >= RAM_START && addr <= RAM_END {
			return c.ram[val]
		}
		return c.rom[val]
	}

	return byte(val)
}

func (c *Cartridge) Write(addr uint16, data byte) {
	val, isAddr := c.mbc.Write(addr, data)
	if isAddr && addr >= RAM_START && addr <= RAM_END {
		c.ram[val] = data
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

func getMBC(header *Header) (MBC, *rtc.State, error) {
	if header.CartType == ROM_ONLY {
		return &RomOnly{}, nil, nil
	} else if header.CartType == MBC_1 {
		return NewMBC1(header.RomSize, header.RamSize), nil, nil
	} else if header.CartType >= MBC_3_START && header.CartType <= MBC_3_END {
		rtcState := rtc.NewState()
		return NewMBC3(header.RomSize, header.RamSize, rtcState), rtcState, nil
	}

	return nil, nil, fmt.Errorf("cart type %d not yet implemented", header.CartType)
}
