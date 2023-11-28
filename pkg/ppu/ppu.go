package ppu

import (
	"github.com/siliconandsolder/go-boy/pkg/bus"
	"github.com/siliconandsolder/go-boy/pkg/interrupts"
)

const (
	MAX_OAM_SEARCH     = 80
	MAX_PIXEL_TRANSFER = 160
	H_BLANK_END        = 143
	V_BLANK_START      = 144
	V_BLANK_END        = 153
	SCANLINE_END       = 455

	TILE_MAP_START_ZERO  = 0x9800
	TILE_MAP_START_ONE   = 0x9C00
	TILE_DATA_START_ZERO = 0x8000
	TILE_DATA_START_ONE  = 0x8800

	BUFFER_SIZE          = 23040
	MAX_OAMS             = 40
	MAX_SPRITES_PER_LINE = 10
)

type Ppu struct {
	lcdControl  *LcdControl
	lcdStatus   *LcdStatus
	scs         *ScrollStatus
	ly          byte
	lyc         byte
	oams        []*OamObj
	lineSprites []*OamObj
	pixelBuffer []uint32
	bufferReady bool
	pixelIdx    uint16

	dot     uint16
	pixels  byte
	fetcher *Fetcher
	bus     *bus.Bus
}

func NewPPU(bus *bus.Bus) *Ppu {
	lcdc := NewLcdControl()
	scs := NewScrollStatus()
	oamSlice := make([]*OamObj, MAX_OAMS)
	for i := range oamSlice {
		oamSlice[i] = NewOamObj()
	}
	return &Ppu{
		lcdControl:  lcdc,
		lcdStatus:   NewLcdStatus(),
		scs:         scs,
		ly:          0,
		lyc:         0,
		oams:        oamSlice,
		lineSprites: make([]*OamObj, 0, MAX_SPRITES_PER_LINE),
		pixelBuffer: make([]uint32, BUFFER_SIZE),
		bufferReady: false,
		pixelIdx:    0,
		dot:         0,
		pixels:      0,
		fetcher:     newFetcher(bus, lcdc, scs),
		bus:         bus,
	}
}

func (ppu *Ppu) Cycle(cycles byte) ([]uint32, error) {

	ppu.readRegisters()
	for i := byte(0); i < cycles; i++ {
		if ppu.lcdControl.enabled == 0 {
			ppu.bus.SetOamAccessible(true)
			ppu.bus.SetVramAccessible(true)
			return nil, nil
		}

		//if ppu.lcdStatus.lycLYEqual == 1 && ppu.dot == 0 && ppu.lcdStatus.lycStatInterrupt == 1 {
		//	ppu.bus.Write(bus.INTERRUPT_REQUEST, interrupts.LCDSTAT)
		//}

		switch ppu.lcdStatus.mode {
		case OAM_SEARCH:
			// TODO: Search OAM for OBJs whose Y coordinate overlap this line
			// for all OAMs, check for y-coord
			// if match, record in lineSprites
			// break after 10 or end of loop

			//for _, oam := range ppu.oams {
			//	if oam.posY <= (ppu.ly+16) && (ppu.ly+16) < (oam.posY+ppu.lcdControl.objSize) {
			//		ppu.lineSprites = append(ppu.lineSprites, oam)
			//	}
			//	if len(ppu.lineSprites) == MAX_SPRITES_PER_LINE {
			//		break
			//	}
			//}

			if ppu.dot == MAX_OAM_SEARCH {
				ppu.fetcher.reset(ppu.ly)
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
				ppu.fetcher.lineX++
				ppu.fetcher.resetIfWindow()
			}

			if ppu.fetcher.lineX == MAX_PIXEL_TRANSFER {
				ppu.lcdStatus.mode = H_BLANK
				ppu.bus.SetVramAccessible(true)
				ppu.bus.SetOamAccessible(true)
				if ppu.lcdStatus.hBlankStatInterrupt == 1 {
					ppu.bus.Write(bus.INTERRUPT_REQUEST, interrupts.LCDSTAT)
				}
			}
			break
		case H_BLANK:
			// wait and go to OAM search, or do vblank if ly == 144
			if ppu.dot == SCANLINE_END {
				if ppu.ly == H_BLANK_END {
					ppu.lcdStatus.mode = V_BLANK
				} else {
					ppu.lcdStatus.mode = OAM_SEARCH
					ppu.bus.SetOamAccessible(false)
					ppu.loadOams()
				}
			}
			break
		case V_BLANK:
			if ppu.dot == SCANLINE_END && ppu.ly == V_BLANK_END {
				ppu.lcdStatus.mode = OAM_SEARCH
				ppu.lineSprites = make([]*OamObj, 0, MAX_SPRITES_PER_LINE)
				ppu.bus.SetOamAccessible(false)
				ppu.loadOams()
				ppu.pixelIdx = 0
				ppu.bufferReady = true
			}
			break
		}

		ppu.dot++
		if ppu.dot == SCANLINE_END+1 {
			ppu.dot = 0
			ppu.ly++
			if ppu.ly == V_BLANK_END+1 {
				ppu.ly = 0
			}

			ppu.lcdStatus.lycLYEqual = 0
			if ppu.lyc == ppu.ly {
				ppu.lcdStatus.lycLYEqual = 1
				if ppu.lcdStatus.lycStatInterrupt == 1 {
					ppu.bus.Write(bus.INTERRUPT_REQUEST, interrupts.LCDSTAT)
				}
			}

			if ppu.ly == V_BLANK_START {
				ppu.bus.Write(bus.INTERRUPT_REQUEST, interrupts.VBLANK)
				if ppu.lcdStatus.vBlankStatInterrupt == 1 {
					ppu.bus.Write(bus.INTERRUPT_REQUEST, interrupts.LCDSTAT)
				}
			} else if ppu.ly < V_BLANK_START && ppu.lcdStatus.oamStatInterrupt == 1 {
				ppu.bus.Write(bus.INTERRUPT_REQUEST, interrupts.LCDSTAT)
			}
		}
	}

	ppu.writeRegisters()

	if ppu.bufferReady {
		ppu.bufferReady = false
		return ppu.pixelBuffer, nil
	}

	return nil, nil
}

func (ppu *Ppu) readRegisters() {
	lcdControlVal := ppu.bus.Read(bus.LCD_CTRL_ADDRESS)
	ppu.lcdControl.enabled = lcdControlVal >> 7 & 1
	ppu.lcdControl.wTileMapArea = lcdControlVal >> 6 & 1
	ppu.lcdControl.windowEnabled = lcdControlVal >> 5 & 1
	ppu.lcdControl.tileDataArea = lcdControlVal >> 4 & 1
	ppu.lcdControl.bgTileMapArea = lcdControlVal >> 3 & 1
	ppu.lcdControl.objSize = lcdControlVal >> 2 & 1
	ppu.lcdControl.objEnabled = lcdControlVal >> 1 & 1
	ppu.lcdControl.bgWindowEnabled = lcdControlVal & 1

	lcdStatusVal := ppu.bus.Read(bus.LCD_STAT_ADDRESS)
	ppu.lcdStatus.lycStatInterrupt = lcdStatusVal >> 6 & 1
	ppu.lcdStatus.oamStatInterrupt = lcdStatusVal >> 5 & 1
	ppu.lcdStatus.vBlankStatInterrupt = lcdStatusVal >> 4 & 1
	ppu.lcdStatus.hBlankStatInterrupt = lcdStatusVal >> 3 & 1

	ppu.scs.scy = ppu.bus.Read(bus.SCY_ADDRESS)
	ppu.scs.scx = ppu.bus.Read(bus.SCX_ADDRESS)

	ppu.lyc = ppu.bus.Read(bus.LCD_LY_ADDRESS)

	ppu.lcdStatus.lycLYEqual = 0
	if ppu.ly == ppu.lyc {
		ppu.lcdStatus.lycLYEqual = 1
	}
}

func (ppu *Ppu) writeRegisters() {
	var lcdStat byte = 0
	lcdStat |= ppu.lcdStatus.lycStatInterrupt << 6
	lcdStat |= ppu.lcdStatus.oamStatInterrupt << 5
	lcdStat |= ppu.lcdStatus.vBlankStatInterrupt << 4
	lcdStat |= ppu.lcdStatus.hBlankStatInterrupt << 3
	lcdStat |= ppu.lcdStatus.lycLYEqual << 2
	lcdStat |= ppu.lcdStatus.mode
	ppu.bus.Write(bus.LCD_STAT_ADDRESS, lcdStat)

	ppu.bus.Write(bus.LCD_Y_ADDRESS, ppu.ly)
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
