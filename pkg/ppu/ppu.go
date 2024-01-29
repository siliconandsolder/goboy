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
	TILE_DATA_START_TWO  = 0x9000

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

	dot       uint16
	pixels    byte
	fifo      *PixelFIFO
	bgFetcher *BgFetcher
	bus       *bus.Bus
}

func NewPPU(bus *bus.Bus) *Ppu {
	lcdc := NewLcdControl()
	scs := NewScrollStatus()
	oamSlice := make([]*OamObj, MAX_OAMS)
	for i := range oamSlice {
		oamSlice[i] = NewOamObj()
	}
	fifo := newFIFO()

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
		fifo:        fifo,
		bgFetcher:   newBgFetcher(bus, lcdc, scs, fifo),
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

			for _, oam := range ppu.oams {
				if oam.posY <= (ppu.ly+16) && (ppu.ly+16) < (oam.posY+ppu.lcdControl.objSize) {
					ppu.lineSprites = append(ppu.lineSprites, oam)
				}
				if len(ppu.lineSprites) == MAX_SPRITES_PER_LINE {
					break
				}
			}

			if ppu.dot == MAX_OAM_SEARCH {
				ppu.bgFetcher.reset(ppu.ly)
				ppu.lcdStatus.mode = PIXEL_TRANSFER
				ppu.bus.SetVramAccessible(false)
			}
			break
		case PIXEL_TRANSFER:
			// send pixels to display
			ppu.bgFetcher.cycle()

			//if ppu.bgFetcher.lineX == 0 && !ppu.fifo.isEmpty() {
			//	for i := byte(0); i < ppu.scs.scx; i++ {
			//		ppu.fifo.pop()
			//	}
			//}

			if !ppu.fifo.isEmpty() {
				if ppu.bgFetcher.lineX == 0 {
					for i := byte(0); i < ppu.scs.scx%8; i++ {
						ppu.fifo.pop()
					}
				}
				if colour, popped := ppu.fifo.pop(); popped {
					ppu.pixelBuffer[ppu.pixelIdx] = getColour(colour)
					ppu.pixelIdx++
				}
				ppu.bgFetcher.lineX++
				ppu.bgFetcher.resetIfWindow()
			}

			//if colour, popped := ppu.fifo.pop(); popped {
			//	ppu.pixelBuffer[ppu.pixelIdx] = getColour(colour)
			//	ppu.pixelIdx++
			//}

			if ppu.bgFetcher.lineX == MAX_PIXEL_TRANSFER {
				ppu.lcdStatus.mode = H_BLANK
				//ppu.renderBackground()
				ppu.bgFetcher.lineX = 0
				ppu.bus.SetVramAccessible(true)
				ppu.bus.SetOamAccessible(true)
				if ppu.lcdStatus.hBlankStatInterrupt == 1 {
					ppu.bus.ToggleInterrupt(interrupts.LCDSTAT)
				}
			}
			break
		case H_BLANK:
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
					ppu.bus.ToggleInterrupt(interrupts.LCDSTAT)
				}
			}

			if ppu.ly == V_BLANK_START {
				ppu.bus.ToggleInterrupt(interrupts.VBLANK)
				if ppu.lcdStatus.vBlankStatInterrupt == 1 {
					ppu.bus.ToggleInterrupt(interrupts.LCDSTAT)
				}
			} else if ppu.ly < V_BLANK_START && ppu.lcdStatus.oamStatInterrupt == 1 {
				ppu.bus.ToggleInterrupt(interrupts.LCDSTAT)
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
	ppu.lcdControl.bgWindowPriority = lcdControlVal & 1

	lcdStatusVal := ppu.bus.Read(bus.LCD_STAT_ADDRESS)
	ppu.lcdStatus.lycStatInterrupt = lcdStatusVal >> 6 & 1
	ppu.lcdStatus.oamStatInterrupt = lcdStatusVal >> 5 & 1
	ppu.lcdStatus.vBlankStatInterrupt = lcdStatusVal >> 4 & 1
	ppu.lcdStatus.hBlankStatInterrupt = lcdStatusVal >> 3 & 1

	ppu.scs.scy = ppu.bus.Read(bus.SCY_ADDRESS)
	ppu.scs.scx = ppu.bus.Read(bus.SCX_ADDRESS)
	ppu.scs.wy = ppu.bus.Read(bus.WY_ADDRESS)
	ppu.scs.wx = ppu.bus.Read(bus.WX_ADDRESS)

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

func (ppu *Ppu) renderBackground() {
	var tileData uint16 = 0
	var bgMem uint16 = 0
	inWindow := ppu.lcdControl.windowEnabled == 1 && ppu.scs.wy <= ppu.ly
	unsigned := true

	if ppu.lcdControl.tileDataArea == 1 {
		tileData = TILE_DATA_START_ZERO
	} else {
		tileData = TILE_DATA_START_ONE
		unsigned = false
	}

	if !inWindow {
		if ppu.lcdControl.bgTileMapArea == 1 {
			bgMem = TILE_MAP_START_ONE
		} else {
			bgMem = TILE_MAP_START_ZERO
		}
	} else {
		if ppu.lcdControl.wTileMapArea == 1 {
			bgMem = TILE_MAP_START_ONE
		} else {
			bgMem = TILE_MAP_START_ZERO
		}
	}

	var yPos byte = 0

	if !inWindow {
		yPos = ppu.scs.scy + ppu.ly
	} else {
		yPos = ppu.ly - ppu.scs.wy
	}

	var tileRow = uint16(yPos/8) * 32

	var pixel byte
	for pixel = 0; pixel < 160; pixel++ {

		var xPos byte
		if inWindow && pixel >= ppu.scs.wx {
			xPos = pixel - ppu.scs.wx
		} else {
			xPos = pixel + ppu.scs.scx
		}

		var tileCol = uint16(xPos / 8)
		var tileNum int16

		tileAddr := bgMem + tileRow + tileCol
		if unsigned {
			tileNum = int16(ppu.bus.PpuReadVram(tileAddr))
		} else {
			tileNum = int16(int8(ppu.bus.PpuReadVram(tileAddr)))
		}

		tileLoc := tileData

		if unsigned {
			tileLoc += uint16(tileNum * 16)
		} else {
			tileLoc = uint16(int32(tileLoc) + int32((tileNum+128)*16))
		}

		line := (yPos % 8) * 2
		dataLow := ppu.bus.PpuReadVram(tileLoc + uint16(line))
		dataHigh := ppu.bus.PpuReadVram(tileLoc + uint16(line) + 1)

		colourBit := byte(int8((xPos%8)-7) * -1)
		colourNum := ((dataHigh>>colourBit)&1)<<1 | ((dataLow >> colourBit) & 1)

		ppu.pixelBuffer[ppu.pixelIdx] = getColour(colourNum)
		ppu.pixelIdx++
	}
}
