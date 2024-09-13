package bus

import (
	"fmt"
	"github.com/siliconandsolder/go-boy/pkg/audio"
	"github.com/siliconandsolder/go-boy/pkg/cartridge"
	"github.com/siliconandsolder/go-boy/pkg/controller"
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

	GLOBAL_MASTER_CONTROL = 0xFF26
	GLOBAL_SOUND_PAN      = 0xFF25
	GLOBAL_MASTER_VOLUME  = 0xFF24

	CHANNEL_ONE_SWEEP       = 0xFF10
	CHANNEL_ONE_LENGTH      = 0xFF11
	CHANNEL_ONE_VOLUME      = 0xFF12
	CHANNEL_ONE_PERIOD_LOW  = 0xFF13
	CHANNEL_ONE_PERIOD_HIGH = 0xFF14

	CHANNEL_TWO_LENGTH      = 0xFF16
	CHANNEL_TWO_VOLUME      = 0xFF17
	CHANNEL_TWO_PERIOD_LOW  = 0xFF18
	CHANNEL_TWO_PERIOD_HIGH = 0xFF19

	CHANNEL_THREE_DAC         = 0xFF1A
	CHANNEL_THREE_LENGTH      = 0xFF1B
	CHANNEL_THREE_OUTPUT      = 0xFF1C
	CHANNEL_THREE_PERIOD_LOW  = 0xFF1D
	CHANNEL_THREE_PERIOD_HIGH = 0xFF1E
	CHANNEL_THREE_WAVE_START  = 0xFF30
	CHANNEL_THREE_WAVE_END    = 0xFF3F

	CHANNEL_FOUR_LENGTH  = 0xFF20
	CHANNEL_FOUR_VOLUME  = 0xFF21
	CHANNEL_FOUR_FREQ    = 0xFF22
	CHANNEL_FOUR_CONTROL = 0xFF23

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
	controller     *controller.Controller
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

	soundChip *audio.SoundChip
}

func NewBus(cart *cartridge.Cartridge, manager *interrupts.Manager, c *controller.Controller) *Bus {
	return &Bus{
		cart:           cart,
		manager:        manager,
		internalRam:    make([]byte, 8192),
		videoRam:       make([]byte, 8192), // bitshift to upper bank
		highRam:        make([]byte, 127),
		oam:            make([]byte, 160),
		dmaSource:      0,
		controller:     c,
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
	case CONTROLLER:
		bus.controller.SetButtonSelectors(value)
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
	case GLOBAL_MASTER_CONTROL:
		bus.soundChip.SetMasterControl(value)
	case GLOBAL_SOUND_PAN:
		bus.soundChip.Global.Panning = value
	case GLOBAL_MASTER_VOLUME:
		bus.soundChip.SetMasterVolume(value)
	case CHANNEL_ONE_SWEEP:
		bus.soundChip.SetPulse1Sweep(value)
	case CHANNEL_ONE_LENGTH:
		bus.soundChip.SetPulse1LengthDuty(value)
	case CHANNEL_ONE_VOLUME:
		bus.soundChip.SetPulse1VolumeEnvelope(value)
	case CHANNEL_ONE_PERIOD_LOW:
		bus.soundChip.SetPulse1PeriodLow(value)
	case CHANNEL_ONE_PERIOD_HIGH:
		bus.soundChip.SetPulse1PeriodHigh(value)
	case CHANNEL_TWO_LENGTH:
		bus.soundChip.SetPulse2LengthDuty(value)
	case CHANNEL_TWO_VOLUME:
		bus.soundChip.SetPulse2VolumeEnvelope(value)
	case CHANNEL_TWO_PERIOD_LOW:
		bus.soundChip.SetPulse2PeriodLow(value)
	case CHANNEL_TWO_PERIOD_HIGH:
		bus.soundChip.SetPulse2PeriodHigh(value)
	case CHANNEL_THREE_DAC:
		bus.soundChip.SetWaveDAC(value)
	case CHANNEL_THREE_LENGTH:
		bus.soundChip.SetWaveLengthTimer(value)
	case CHANNEL_THREE_OUTPUT:
		bus.soundChip.SetWaveOutput(value)
	case CHANNEL_THREE_PERIOD_LOW:
		bus.soundChip.SetWavePeriodLow(value)
	case CHANNEL_THREE_PERIOD_HIGH:
		bus.soundChip.SetWavePeriodHigh(value)
	case CHANNEL_FOUR_LENGTH:
		bus.soundChip.SetNoiseLengthTimer(value)
	case CHANNEL_FOUR_VOLUME:
		bus.soundChip.SetNoiseVolumeEnvelope(value)
	case CHANNEL_FOUR_FREQ:
		bus.soundChip.SetNoiseFreqRandomness(value)
	case CHANNEL_FOUR_CONTROL:
		bus.soundChip.SetNoiseControl(value)
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
	} else if addr >= CHANNEL_THREE_WAVE_START && addr <= CHANNEL_THREE_WAVE_END {
		bus.soundChip.SetWaveRAM(addr-CHANNEL_THREE_WAVE_START, value)
	}
}
func (bus *Bus) Read(addr uint16) byte {
	switch addr {
	case CONTROLLER:
		return bus.controller.GetJoypadValue()
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
	case GLOBAL_MASTER_CONTROL:
		return bus.soundChip.GetMasterControl()
	case GLOBAL_SOUND_PAN:
		return bus.soundChip.Global.Panning
	case GLOBAL_MASTER_VOLUME:
		return bus.soundChip.GetMasterVolume()
	case CHANNEL_ONE_SWEEP:
		return bus.soundChip.GetPulse1Sweep()
	case CHANNEL_ONE_LENGTH:
		return bus.soundChip.GetPulse1LengthDuty()
	case CHANNEL_ONE_VOLUME:
		return bus.soundChip.GetPulse1VolumeEnvelope()
	case CHANNEL_ONE_PERIOD_HIGH:
		return bus.soundChip.GetPulse1PeriodHigh()
	case CHANNEL_TWO_LENGTH:
		return bus.soundChip.GetPulse2LengthDuty()
	case CHANNEL_TWO_VOLUME:
		return bus.soundChip.GetPulse2VolumeEnvelope()
	case CHANNEL_TWO_PERIOD_HIGH:
		return bus.soundChip.GetPulse2PeriodHigh()
	case CHANNEL_THREE_DAC:
		return bus.soundChip.GetWaveDAC()
	case CHANNEL_THREE_OUTPUT:
		return bus.soundChip.GetWaveOutput()
	case CHANNEL_THREE_PERIOD_HIGH:
		return bus.soundChip.GetWaveLengthEnable()
	case CHANNEL_FOUR_VOLUME:
		return bus.soundChip.GetNoiseVolumeEnvelope()
	case CHANNEL_FOUR_FREQ:
		return bus.soundChip.GetNoiseFreqRandomness()
	case CHANNEL_FOUR_CONTROL:
		return bus.soundChip.GetNoiseControl()
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
	} else if addr >= CHANNEL_THREE_WAVE_START && addr <= CHANNEL_THREE_WAVE_END {
		return bus.soundChip.GetWaveRAM(addr - CHANNEL_THREE_WAVE_START)
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
