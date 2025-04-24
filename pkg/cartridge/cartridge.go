package cartridge

import (
	"encoding/json"
	"fmt"
	"github.com/siliconandsolder/go-boy/pkg/cartridge/rtc"
	"os"
	"slices"
	"strings"
)

const (
	ROM_ONLY    = 0x00
	MBC_1_START = 0x01
	MBC_1_END   = 0x03
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

var batteryCartridges = []byte{0x03, 0x06, 0x09, 0x0D, 0x0F, 0x10, 0x13, 0x1B, 0x1E, 0x22, 0xFF}

type Cartridge struct {
	Title      string
	header     *Header
	mbc        MBC
	rom        []byte
	ram        []byte
	state      *rtc.State
	hasBattery bool
}

type SaveFile struct {
	Sram        []byte
	RtcSnapshot *rtc.StateSnapshot
}

func NewCartridge(file []byte) *Cartridge {
	header := NewHeader(file[HEADER_START:])

	if err := verifyChecksum(header.HeaderChecksum, file[CHECKSUM_LOWER_BOUND:CHECKSUM_UPPER_BOUND]); err != nil {
		panic(err)
	}

	mapper, state, err := getMBC(header)
	if err != nil {
		panic(err)
	}

	rom := make([]byte, header.RomSize.Size)
	ram := make([]byte, header.RamSize.Size)

	rom = file

	return &Cartridge{
		Title:      header.Title,
		header:     header,
		mbc:        mapper,
		rom:        rom,
		ram:        ram,
		state:      state,
		hasBattery: slices.Contains(batteryCartridges, header.CartType),
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

func (c *Cartridge) SaveRAMToFile() {
	if c.hasBattery {
		if c.Title == "" {
			return
		}

		sram := make([]byte, c.header.RamSize.Size)
		copy(sram, c.ram)

		var snapshot *rtc.StateSnapshot = nil
		if c.state != nil {
			snapshot = c.state.GetSnapshot()
		}

		save := SaveFile{
			Sram:        sram,
			RtcSnapshot: snapshot,
		}

		saveJson, err := json.Marshal(save)
		if err != nil {
			panic(err)
		}

		if _, err := os.Stat("saves"); os.IsNotExist(err) {
			err = os.Mkdir("saves", 0777)
			if err != nil {
				panic(err)
			}
		}

		saveTitle := fmt.Sprintf("%s.sav", strings.ToLower(strings.ReplaceAll(c.Title, " ", "_")))
		err = os.WriteFile("saves/"+saveTitle, saveJson, 0777)
		if err != nil {
			panic(err)
		}
	}
}

func (c *Cartridge) LoadRAMFromFile() {
	if c.hasBattery {
		if c.Title == "" {
			return
		}

		saveTitle := fmt.Sprintf("%s.sav", strings.ToLower(strings.ReplaceAll(c.Title, " ", "_")))
		saveData, err := os.ReadFile("saves/" + saveTitle)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("no save file for: %s\n", c.Title)
				return
			} else {
				panic(err)
			}
		}

		save := SaveFile{}
		err = json.Unmarshal(saveData, &save)
		if err != nil {
			fmt.Printf("could not read save: %v\n", err)
			return
		}

		copy(c.ram, save.Sram)

		if save.RtcSnapshot != nil {
			c.state.FromSnapshot(save.RtcSnapshot)
		}
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
	} else if header.CartType >= MBC_1_START && header.CartType <= MBC_1_END {
		return NewMBC1(header.RomSize, header.RamSize), nil, nil
	} else if header.CartType >= MBC_3_START && header.CartType <= MBC_3_END {
		rtcState := rtc.NewState()
		return NewMBC3(header.RomSize, header.RamSize, rtcState), rtcState, nil
	}

	return nil, nil, fmt.Errorf("cart type %d not yet implemented", header.CartType)
}
