package ppu

import (
	"cmp"
	"github.com/siliconandsolder/go-boy/pkg/bus"
	"github.com/siliconandsolder/go-boy/pkg/interrupts"
	"slices"
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

type FetcherState byte

const (
	ReadTileID FetcherState = iota
	ReadTileData0
	ReadTileData1
	PushToFIFO
)

type Ppu struct {
	lcdControl  *LcdControl
	lcdStatus   *LcdStatus
	scs         *ScrollStatus
	ly          byte
	lyc         byte
	x           byte
	oams        []*OamObj
	lineSprites []*OamObj
	pixelBuffer []uint32
	bufferReady bool
	pixelIdx    uint16

	dot            uint16
	pixels         byte
	bgFetcher      *BgFetcher
	fgFetcher      *SpriteFetcher
	fetchingSprite bool
	shouldCycle    bool
	bus            *bus.Bus
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
		x:           0,
		oams:        oamSlice,
		lineSprites: make([]*OamObj, 0, MAX_SPRITES_PER_LINE),
		pixelBuffer: make([]uint32, BUFFER_SIZE),
		bufferReady: false,
		pixelIdx:    0,
		dot:         0,
		pixels:      0,
		bgFetcher:   newBgFetcher(bus, lcdc, scs),
		fgFetcher:   newSpriteFetcher(bus, lcdc),
		bus:         bus,
	}
}

func (ppu *Ppu) Cycle(cycles byte) ([]uint32, error) {

	ppu.readRegisters()
	for i := byte(0); i < cycles; i++ {
		if ppu.lcdControl.enabled == 0 {
			ppu.bus.SetOamAccessible(true)
			ppu.bus.SetVramAccessible(true)
			ppu.lcdStatus.mode = 0
			ppu.pixelIdx = 0
			ppu.ly = 0
			ppu.bufferReady = false
			break // wait for all cycles to complete
		}

		switch ppu.lcdStatus.mode {
		case OAM_SEARCH:
			if ppu.dot == MAX_OAM_SEARCH {
				ppu.lineSprites = make([]*OamObj, 0, MAX_SPRITES_PER_LINE)
				var spriteSize byte
				if ppu.lcdControl.objSize == 0 {
					spriteSize = 8
				} else {
					spriteSize = 16
				}
				for idx, oam := range ppu.oams {
					if oam.posY <= (ppu.ly+16) && (ppu.ly+16) < (oam.posY+spriteSize) {
						oam.idx = byte(idx)
						ppu.lineSprites = append(ppu.lineSprites, oam)
					}
					if len(ppu.lineSprites) == MAX_SPRITES_PER_LINE {
						break
					}
				}

				slices.SortStableFunc(ppu.lineSprites, func(a, b *OamObj) int {
					xPriority := cmp.Compare(a.posX, b.posX)
					if xPriority != 0 {
						return xPriority
					} else {
						return cmp.Compare(a.idx, b.idx)
					}
				})

				ppu.bgFetcher.reset(ppu.x, ppu.ly)
				ppu.fgFetcher.reset(ppu.ly)
				ppu.lcdStatus.mode = PIXEL_TRANSFER
				ppu.bus.SetVramAccessible(false)
			}
			break
		case PIXEL_TRANSFER:
			if ppu.fgFetcher.spriteToFetch != nil {
				ppu.fgFetcher.cycle(ppu.shouldCycle)
			} else if sprite := ppu.checkForSprite(); sprite != nil {
				ppu.fgFetcher.spriteToFetch = sprite
			}

			ppu.bgFetcher.cycle(ppu.shouldCycle)
			ppu.shouldCycle = !ppu.shouldCycle

			if !ppu.bgFetcher.fifo.isEmpty() && ppu.fgFetcher.spriteToFetch == nil {
				if ppu.x == 0 {
					for i := byte(0); i < ppu.scs.scx%8; i++ {
						ppu.bgFetcher.fifo.pop()
					}
				}
				if !ppu.fgFetcher.fifo.isEmpty() {
					bgPixel := ppu.bgFetcher.fifo.pop()
					fgPixel := ppu.fgFetcher.fifo.pop()

					if ppu.lcdControl.bgWindowEnabled == 0 {
						bgPixel.colourNum = 0
					}
					if ppu.lcdControl.objEnabled == 0 {
						fgPixel.colourNum = 0
					}

					if fgPixel.colourNum == 0 || (fgPixel.priority == 1 && bgPixel.colourNum != 0) {
						colour := (ppu.bus.Read(bgPixel.paletteAddr) & (0x3 << (bgPixel.colourNum * 2))) >> (bgPixel.colourNum * 2)
						ppu.pixelBuffer[ppu.pixelIdx] = getColour(colour)
						ppu.pixelIdx++
					} else {
						colour := (ppu.bus.Read(fgPixel.paletteAddr) & (0x3 << (fgPixel.colourNum * 2))) >> (fgPixel.colourNum * 2)
						ppu.pixelBuffer[ppu.pixelIdx] = getColour(colour)
						ppu.pixelIdx++
					}
				} else {
					pixel := ppu.bgFetcher.fifo.pop()
					if ppu.lcdControl.bgWindowEnabled == 0 {
						pixel.colourNum = 0
					}
					colour := (ppu.bus.Read(pixel.paletteAddr) & (0x3 << (pixel.colourNum * 2))) >> (pixel.colourNum * 2)
					ppu.pixelBuffer[ppu.pixelIdx] = getColour(colour)
					ppu.pixelIdx++
				}

				ppu.x++
				ppu.bgFetcher.resetIfWindow(ppu.x, ppu.ly)
			}

			if ppu.x == MAX_PIXEL_TRANSFER {
				ppu.lcdStatus.mode = H_BLANK
				ppu.x = 0
				ppu.bus.SetVramAccessible(true)
				ppu.bus.SetOamAccessible(true)
				if ppu.lcdStatus.hBlankStatInterrupt == 1 {
					ppu.bus.ToggleInterrupt(interrupts.LCDSTAT)
				}
			}
			break
		case H_BLANK:
			if ppu.bgFetcher.inWindow {
				ppu.bgFetcher.windowCounter++
				ppu.bgFetcher.inWindow = false
			}
			if ppu.dot == SCANLINE_END {
				if ppu.ly == H_BLANK_END {
					ppu.bgFetcher.windowCounter = 0
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
	ppu.lcdControl.bgWindowEnabled = lcdControlVal & 1

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
		oamY := ppu.bus.PpuReadOam(uint16(i * 4))
		oamX := ppu.bus.PpuReadOam(uint16(i*4 + 1))
		oamIdx := ppu.bus.PpuReadOam(uint16(i*4 + 2))
		oamAttrs := ppu.bus.PpuReadOam(uint16(i*4 + 3))

		ppu.oams[i] = &OamObj{
			posX:    oamX,
			posY:    oamY,
			tileNum: oamIdx,
			attributes: OamAttributes{
				priority: oamAttrs >> 7 & 1,
				yFlip:    oamAttrs >> 6 & 1,
				xFlip:    oamAttrs >> 5 & 1,
				palette:  oamAttrs >> 4 & 1,
			},
		}
	}
}

func (ppu *Ppu) checkForSprite() *OamObj {
	for _, sprite := range ppu.lineSprites {
		if sprite.posX <= ppu.x+8 {
			ppu.lineSprites = slices.Delete(ppu.lineSprites, 0, 1)
			return sprite
		}
	}

	return nil
}
