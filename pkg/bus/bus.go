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

	CONTROLLER = 0xFF00

	DMA_SOURCE = 0xFF46

	OAM_START = 0xFE00
	OAM_END   = 0xFE9F

	HIGH_RAM_START = 0xFF80
	HIGH_RAM_END   = 0xFFFE

	// TODO: implement PPU registers
	LCD_CTRL_ADDRESS  = 0xFF40
	LCD_STAT_ADDRESS  = 0xFF41
	SCY_ADDRESS       = 0xFF42
	SCX_ADDRESS       = 0xFF43
	LCD_Y_ADDRESS     = 0xFF44
	LCD_LY_ADDRESS    = 0xFF45
	DMA_TRANSFER      = 0xFF46
	BG_PALETTE        = 0xFF47
	FG_PALETTE_ZERO   = 0xFF48
	FG_PALETTE_ONE    = 0xFF49
	WY_ADDRESS        = 0xFF4A
	WX_ADDRESS        = 0xFF4B
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
	dmaSource      byte
	lcdCtrl        byte
	lcdStat        byte
	lcdY           byte
	lcdLy          byte
	scy            byte
	scx            byte
	wy             byte
	wx             byte
	serialByte     byte
	bgPalette      byte
	fgPaletteZero  byte
	fgPaletteOne   byte
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
		dmaSource:      0,
		serialByte:     0,
		lcdCtrl:        0x95,
		lcdStat:        0x85,
		scy:            0,
		scx:            0,
		wy:             0,
		wx:             0,
		vramAccessible: true,
		oamAccessible:  true,
	}
}

func (bus *Bus) Write(addr uint16, value byte) {

	switch addr {
	case INTERRUPT_REQUEST:
		bus.manager.SetInterruptRequest(value)
	case INTERRUPT_ENABLE:
		bus.manager.SetInterruptEnable(value)
	case SERIAL_TRANSFER_DATA:
		bus.serialByte = value
	case SERIAL_TRANSFER_CONTROL:
		if value == 0x81 {
			fmt.Print(fmt.Sprintf("%c", bus.serialByte))
		}
	case DMA_SOURCE:
		bus.dmaSource = value
	case LCD_CTRL_ADDRESS:
		bus.lcdCtrl = value
	case LCD_STAT_ADDRESS:
		bus.lcdStat = value
	case SCY_ADDRESS:
		bus.scy = value
	case SCX_ADDRESS:
		bus.scx = value
	case LCD_Y_ADDRESS:
		bus.lcdY = value
	case LCD_LY_ADDRESS:
		bus.lcdLy = value
	case BG_PALETTE:
		bus.bgPalette = value
	case FG_PALETTE_ZERO:
		bus.fgPaletteZero = value
	case FG_PALETTE_ONE:
		bus.fgPaletteOne = value
	case WY_ADDRESS:
		bus.wy = value
	case WX_ADDRESS:
		bus.wx = value
	}

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
	}
}
func (bus *Bus) Read(addr uint16) byte {
	switch addr {
	case CONTROLLER:
		return 0xFF
	case INTERRUPT_REQUEST:
		return bus.manager.GetInterruptRequests()
	case INTERRUPT_ENABLE:
		return bus.manager.GetEnabledInterrupts()
	case LCD_CTRL_ADDRESS:
		return bus.lcdCtrl
	case LCD_STAT_ADDRESS:
		return bus.lcdStat
	case SCY_ADDRESS:
		return bus.scy
	case SCX_ADDRESS:
		return bus.scx
	case DMA_SOURCE:
		return bus.dmaSource
	case LCD_Y_ADDRESS:
		return bus.lcdY
	case LCD_LY_ADDRESS:
		return bus.lcdLy
	case BG_PALETTE:
		return bus.bgPalette
	case FG_PALETTE_ZERO:
		return bus.fgPaletteZero
	case FG_PALETTE_ONE:
		return bus.fgPaletteOne
	case WY_ADDRESS:
		return bus.wy
	case WX_ADDRESS:
		return bus.wx
	}

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
	} else if addr >= HIGH_RAM_START && addr <= HIGH_RAM_END {
		return bus.highRam[addr-HIGH_RAM_START]
	}

	return 0
}

func (bus *Bus) PpuReadVram(addr uint16) byte {
	return bus.videoRam[addr-VRAM_START]
}

func (bus *Bus) PpuReadOam(addr uint16) byte {
	return bus.oam[addr]
}

func (bus *Bus) SetVramAccessible(access bool) {
	bus.vramAccessible = access
}

func (bus *Bus) SetOamAccessible(access bool) {
	bus.oamAccessible = access
}

func (bus *Bus) ToggleInterrupt(val byte) {
	bus.manager.ToggleInterruptRequest(val)
}
