package ppu

import "github.com/siliconandsolder/go-boy/pkg/bus"

type FetcherState byte

const (
	ReadTileID FetcherState = iota
	ReadTileData0
	ReadTileData1
	PushToFIFO
)

type BgFetcher struct {
	fifo        *PixelFIFO
	bus         *bus.Bus
	lcdc        *LcdControl
	scs         *ScrollStatus
	shouldCycle bool
	state       FetcherState
	tileData    []byte

	tileIdx     byte
	tileId      byte
	mapAddr     uint16
	tileLine    byte
	tileOffset  int32
	lineX       byte
	pixelX      byte
	tileX       byte
	lineY       byte
	pixelY      byte
	tileY       byte
	inWindow    bool
	isFirstTile bool
}

func newBgFetcher(bus *bus.Bus, lcdc *LcdControl, scs *ScrollStatus, fifo *PixelFIFO) *BgFetcher {
	return &BgFetcher{
		fifo:        fifo,
		bus:         bus,
		lcdc:        lcdc,
		scs:         scs,
		shouldCycle: false,
		state:       0,
		tileData:    make([]byte, 8),
		tileId:      0,
		tileOffset:  0,
		mapAddr:     0,
		tileLine:    0,
		lineX:       0,
		pixelX:      0,
		tileX:       0,
		lineY:       0,
		pixelY:      0,
		tileY:       0,
		inWindow:    false,
		isFirstTile: false,
	}
}

func (f *BgFetcher) reset(lineY byte) {
	f.lineY = lineY
	f.state = ReadTileID

	if f.lcdc.windowEnabled == 1 && f.lineY >= f.scs.wy {
		f.pixelY = lineY - f.scs.wy
	} else {
		f.pixelY = f.scs.scy + lineY
	}
	f.tileY = (f.pixelY >> 3) & 31
	if f.lcdc.windowEnabled == 1 && f.lineY >= f.scs.wy && f.lineX >= (f.scs.wx-7) {
		f.inWindow = true
	} else {
		f.inWindow = false
	}

	if f.inWindow {
		f.pixelX = f.lineX - (f.scs.wx - 7)
	} else {
		f.pixelX = f.scs.scx + f.lineX
	}

	f.tileX = 0
	f.fifo.clear()
}

func (f *BgFetcher) cycle() {
	if !f.shouldCycle {
		f.shouldCycle = true
		return
	}

	f.shouldCycle = false

	switch f.state {
	case ReadTileID:
		f.mapAddr = f.getTileMapBase() + uint16(f.tileY)*32 + uint16(((f.pixelX>>3)+f.tileX)&31)

		f.tileId = f.bus.PpuReadVram(f.mapAddr)
		//if f.lcdc.tileDataArea == 1 {
		//	f.tileOffset = int32(int16(f.tileId))
		//} else {
		//	f.tileOffset = int32(int16(int8(f.tileId)) + 128)
		//}
		//f.tileOffset *= 16
		f.state = ReadTileData0
		break
	case ReadTileData0:
		f.readTileLine(false)
		f.state = ReadTileData1
		break
	case ReadTileData1:
		f.readTileLine(true)
		f.state = PushToFIFO
		break
	case PushToFIFO:
		if f.fifo.size <= 8 {
			for i := 7; i >= 0; i-- {
				if err := f.fifo.push(f.tileData[i]); err != nil {
					panic(err)
				}
			}
			f.tileX++
			f.state = ReadTileID
		}
		break
	}
}

func (f *BgFetcher) readTileLine(isHigh bool) {
	// get tile data base
	addr := f.getTileDataBase() + uint16(f.tileId)*16 + uint16(f.pixelY%8)<<1

	if isHigh {
		addr++
	}
	data := f.bus.PpuReadVram(addr)

	for bitPos := byte(0); bitPos < 8; bitPos++ {
		if isHigh {
			f.tileData[bitPos] |= ((data >> bitPos) & 1) << 1
		} else {
			f.tileData[bitPos] = (data >> bitPos) & 1
		}
	}

	if isHigh && f.isFirstTile {
		f.isFirstTile = false
		f.state = ReadTileID
	}
}

func (f *BgFetcher) resetIfWindow() {
	// stop pushing to FIFO, we reached the window
	realWx := f.scs.wx - 7
	if f.lcdc.windowEnabled == 1 && f.lineY >= f.scs.wy && f.lineX >= realWx {
		f.reset(f.lineY)
	}
}

func (f *BgFetcher) getTileDataBase() uint16 {
	if f.lcdc.tileDataArea == 0 && f.tileId < 128 {
		return TILE_DATA_START_TWO
	} else {
		return TILE_DATA_START_ZERO
	}
}

func (f *BgFetcher) getTileMapBase() uint16 {
	if f.inWindow {
		if f.lcdc.wTileMapArea == 1 {
			return TILE_MAP_START_ONE
		} else {
			return TILE_MAP_START_ZERO
		}
	} else if f.lcdc.bgTileMapArea == 1 {
		return TILE_MAP_START_ONE
	} else {
		return TILE_MAP_START_ZERO
	}
}
