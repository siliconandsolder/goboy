package ppu

import (
	"github.com/siliconandsolder/go-boy/pkg/bus"
	"github.com/siliconandsolder/go-boy/pkg/interrupts"
)

const (
	MAX_OAM_SEARCH     = 80
	MAX_PIXEL_TRANSFER = 160
	V_BLANK_START      = 144
	V_BLANK_END        = 153
	SCANLINE_END       = 456

	TILE_MAP_START_ZERO = 0x9800
	TILE_MAP_START_ONE  = 0x9C00
	TILE_DATA_START     = 0x8000

	LCD_CONTROL_ADDR = 0xFF40
	LCD_STATUS_ADDR  = 0xFF41

	BUFFER_SIZE          = 23040
	MAX_OAMS             = 40
	MAX_SPRITES_PER_LINE = 10
)

type Ppu struct {
	lcdControl  *LcdControl
	lcdStatus   *LcdStatus
	scy         byte
	scx         byte
	ly          byte
	lyc         byte
	wy          byte
	wx          byte
	oams        []*OamObj
	lineSprites []*OamObj
	pixelBuffer []uint32
	pixelIdx    uint16

	dot     uint16
	pixels  byte
	fetcher *Fetcher
	bus     *bus.Bus
}

func NewPPU(bus *bus.Bus) *Ppu {
	return &Ppu{
		lcdControl:  NewLcdControl(),
		lcdStatus:   NewLcdStatus(),
		scy:         0,
		scx:         0,
		ly:          0,
		lyc:         0,
		wy:          0,
		wx:          0,
		oams:        make([]*OamObj, MAX_OAMS),
		lineSprites: make([]*OamObj, 0, MAX_SPRITES_PER_LINE),
		pixelBuffer: make([]uint32, BUFFER_SIZE),
		pixelIdx:    0,
		dot:         0,
		pixels:      0,
		fetcher:     newFetcher(bus),
		bus:         bus,
	}
}

func (ppu *Ppu) Cycle() ([]uint32, error) {
	ppu.refreshRegisters()

	if ppu.ly == ppu.lyc && ppu.lcdStatus.lycLYEqual == 1 {
		ppu.bus.Write(bus.INTERRUPT_REQUEST, interrupts.LCDSTAT)
	}

	switch ppu.lcdStatus.mode {
	case OAM_SEARCH:
		if ppu.dot == 0 && ppu.lcdStatus.oamStatInterrupt == 1 {
			ppu.bus.Write(bus.INTERRUPT_REQUEST, interrupts.LCDSTAT)
		}

		// TODO: Search OAM for OBJs whose Y coordinate overlap this line
		// for all OAMs, check for y-coord
		// if match, record in lineSprites
		// break after 10 or end of loop

		for _, oam := range ppu.oams {
			if oam.posY <= (ppu.ly+16) && (ppu.ly+16) < (oam.posY+ppu.lcdControl.objSize) {
				ppu.lineSprites = append(ppu.lineSprites, oam)
			}
			if len(ppu.lineSprites) == MAX_SPRITES_PER_LINE {
				break
			}
		}

		if ppu.dot == MAX_OAM_SEARCH {
			ppu.pixels = 0
			tileLine := ppu.ly & 7
			tileMapAddr := TILE_MAP_START_ZERO + (uint16(ppu.ly>>3) << 5)
			ppu.fetcher.reset(tileMapAddr, tileLine)
			ppu.lcdStatus.mode = PIXEL_TRANSFER
			ppu.bus.SetVramAccessible(false)
		}
		break
	case PIXEL_TRANSFER:
		// send pixels to display
		ppu.fetcher.cycle()

		if colour, popped := ppu.fetcher.fifo.pop(); popped {
			ppu.pixelBuffer[ppu.pixelIdx] = getColour(colour)
			ppu.pixelIdx++
			ppu.pixels++
		}

		if ppu.pixels == MAX_PIXEL_TRANSFER {
			ppu.lcdStatus.mode = H_BLANK
			ppu.bus.SetVramAccessible(true)
			ppu.bus.SetOamAccessible(true)
		}
		break
	case H_BLANK:
		if ppu.dot == 0 && ppu.lcdStatus.hBlankStatInterrupt == 1 {
			ppu.bus.Write(bus.INTERRUPT_REQUEST, interrupts.LCDSTAT)
		}
		// wait and go to OAM search, or do vblank if ly == 144
		if ppu.dot == SCANLINE_END {
			ppu.dot = 0

			ppu.ly++
			if ppu.ly == V_BLANK_START {
				ppu.lcdStatus.mode = V_BLANK
			} else {
				ppu.lcdStatus.mode = OAM_SEARCH
				ppu.bus.SetOamAccessible(false)
				ppu.loadOams()
			}
			return nil, nil
		}
		break
	case V_BLANK:
		if ppu.ly == V_BLANK_END && ppu.dot == 0 {
			ppu.bus.Write(bus.INTERRUPT_REQUEST, interrupts.VBLANK)

			if ppu.lcdStatus.vBlankStatInterrupt == 1 {
				ppu.bus.Write(bus.INTERRUPT_REQUEST, interrupts.LCDSTAT)
			}
		}
		if ppu.dot == SCANLINE_END {
			ppu.dot = 0

			ppu.ly++
			if ppu.ly == V_BLANK_END {
				ppu.ly = 0
				ppu.lcdStatus.mode = OAM_SEARCH
				ppu.lineSprites = make([]*OamObj, 0, MAX_SPRITES_PER_LINE)
				ppu.bus.SetOamAccessible(false)
				ppu.loadOams()
				ppu.pixelIdx = 0
				return ppu.pixelBuffer, nil
			}
			return nil, nil
		}
		break
	}

	ppu.dot++
	return nil, nil
}

func (ppu *Ppu) refreshRegisters() {
	lcdControlVal := ppu.bus.Read(LCD_CONTROL_ADDR)
	ppu.lcdControl.enabled = lcdControlVal >> 7 & 1
	ppu.lcdControl.wTileMapArea = lcdControlVal >> 6 & 1
	ppu.lcdControl.windowEnabled = lcdControlVal >> 5 & 1
	ppu.lcdControl.tileDataArea = lcdControlVal >> 4 & 1
	ppu.lcdControl.bgTileMapArea = lcdControlVal >> 3 & 1
	ppu.lcdControl.objSize = lcdControlVal >> 2 & 1
	ppu.lcdControl.objEnabled = lcdControlVal >> 1 & 1
	ppu.lcdControl.bgWindowEnabled = lcdControlVal & 1

	lcdStatusVal := ppu.bus.Read(LCD_STATUS_ADDR)
	ppu.lcdStatus.lycStatInterrupt = lcdStatusVal >> 6 & 1
	ppu.lcdStatus.oamStatInterrupt = lcdStatusVal >> 5 & 1
	ppu.lcdStatus.vBlankStatInterrupt = lcdStatusVal >> 4 & 1
	ppu.lcdStatus.hBlankStatInterrupt = lcdStatusVal >> 3 & 1
	if ppu.lyc == ppu.ly {
		ppu.lcdStatus.lycLYEqual = 1
	} else {
		ppu.lcdStatus.lycLYEqual = 0
	}

	ppu.bus.Write(LCD_STATUS_ADDR, (lcdStatusVal&0b11111000)|ppu.lcdStatus.lycLYEqual<<2|ppu.lcdStatus.mode)
}

func (ppu *Ppu) loadOams() {
	for i := 0; i < MAX_OAMS; i++ {
		oamY := ppu.bus.Read(uint16(bus.OAM_START + i*4))
		oamX := ppu.bus.Read(uint16(bus.OAM_START + i*4 + 1))
		oamIdx := ppu.bus.Read(uint16(bus.OAM_START + i*4 + 2))
		oamAttrs := ppu.bus.Read(uint16(bus.OAM_START + i*4 + 3))

		ppu.oams[i] = &OamObj{
			posX:    oamX,
			posY:    oamY,
			tileNum: oamIdx,
			flags:   oamAttrs,
		}
	}
}
