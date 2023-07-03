package bus

import (
	"fmt"
	"github.com/siliconandsolder/go-boy/pkg/cartridge"
	"github.com/siliconandsolder/go-boy/pkg/interrupts"
)

const (
	CART_START = 0x0000
	CART_END   = 0x7FFF

	VRAM_START = 0x8000
	VRAM_END   = 0x9FFF

	INTERNAL_RAM_START = 0xC000
	INTERNAL_RAM_END   = 0xDFFF

	OAM_START = 0xFE00
	OAM_END   = 0xFE9F

	HIGH_RAM_START = 0xFF80
	HIGH_RAM_END   = 0xFFFE

	DMA_TRANSFER      = 0xFF46
	INTERRUPT_REQUEST = 0xFF0F
	INTERRUPT_ENABLE  = 0xFFFF

	SERIAL_TRANSFER_DATA    = 0xFF01
	SERIAL_TRANSFER_CONTROL = 0xFF02
)

type Bus struct {
	cart *cartridge.Cartridge

	internalRam    []byte
	videoRam       []byte
	highRam        []byte
	oam            []byte
	serialByte     byte
	vramAccessible bool
	oamAccessible  bool

	manager *interrupts.Manager
}

func NewBus(cart *cartridge.Cartridge, manager *interrupts.Manager) *Bus {
	return &Bus{
		cart:           cart,
		manager:        manager,
		internalRam:    make([]byte, 8192),
		videoRam:       make([]byte, 8192), // bitshift to upper bank
		highRam:        make([]byte, 127),
		oam:            make([]byte, 160),
		serialByte:     0,
		vramAccessible: true,
		oamAccessible:  true,
	}
}

func (bus *Bus) Write(addr uint16, value byte) {
	if addr <= CART_END {
		bus.cart.Write(addr, value)
	} else if addr >= VRAM_START && addr <= VRAM_END && bus.vramAccessible {
		bus.videoRam[addr-VRAM_START] = value
	} else if addr >= INTERNAL_RAM_START && addr <= INTERNAL_RAM_END {
		bus.internalRam[addr-INTERNAL_RAM_START] = value
	} else if addr >= OAM_START && addr <= OAM_END && bus.oamAccessible {
		bus.oam[addr-OAM_START] = value
	} else if addr >= HIGH_RAM_START && addr <= HIGH_RAM_END {
		bus.highRam[addr-HIGH_RAM_START] = value
	} else if addr == INTERRUPT_REQUEST {
		bus.manager.SetInterruptRequest(value)
	} else if addr == INTERRUPT_ENABLE {
		bus.manager.SetInterruptEnable(value)
	} else if addr == SERIAL_TRANSFER_DATA {
		bus.serialByte = value
	} else if addr == SERIAL_TRANSFER_CONTROL {
		if value == 0x81 {
			fmt.Print(fmt.Sprintf("%c", bus.serialByte))
		}
	}
}
func (bus *Bus) Read(addr uint16) byte {
	if addr <= CART_END {
		return bus.cart.Read(addr)
	} else if addr >= VRAM_START && addr <= VRAM_END {
		if bus.vramAccessible {
			return bus.videoRam[addr-VRAM_START]
		} else {
			return 0xFF
		}
	} else if addr >= INTERNAL_RAM_START && addr <= INTERNAL_RAM_END {
		return bus.internalRam[addr-INTERNAL_RAM_START]
	} else if addr >= OAM_START && addr <= OAM_END {
		if bus.oamAccessible {
			return bus.oam[addr-OAM_START]
		} else {
			return 0xFF
		}
	} else if addr == INTERRUPT_REQUEST {
		return bus.manager.GetInterruptRequest()
	} else if addr >= HIGH_RAM_START && addr <= HIGH_RAM_END {
		return bus.highRam[addr-HIGH_RAM_START]
	}

	return 0
}

func (bus *Bus) SetVramAccessible(access bool) {
	bus.vramAccessible = access
}

func (bus *Bus) SetOamAccessible(access bool) {
	bus.oamAccessible = access
}
