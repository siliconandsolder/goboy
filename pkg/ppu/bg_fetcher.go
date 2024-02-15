package ppu

import "github.com/siliconandsolder/go-boy/pkg/bus"

type BgFetcher struct {
	fifo     *PixelFIFO
	bus      *bus.Bus
	lcdc     *LcdControl
	scs      *ScrollStatus
	state    FetcherState
	tileData []byte

	tileIdx       byte
	tileId        byte
	mapAddr       uint16
	tileLine      byte
	tileOffset    int32
	pixelX        byte
	fetcherX      byte
	pixelY        byte
	tileY         byte
	windowCounter byte
	inWindow      bool
	isFirstTile   bool
}

func newBgFetcher(bus *bus.Bus, lcdc *LcdControl, scs *ScrollStatus) *BgFetcher {
	return &BgFetcher{
		fifo:          newFIFO(true),
		bus:           bus,
		lcdc:          lcdc,
		scs:           scs,
		state:         0,
		tileData:      make([]byte, 8),
		tileId:        0,
		tileOffset:    0,
		mapAddr:       0,
		tileLine:      0,
		pixelX:        0,
		fetcherX:      0,
		pixelY:        0,
		tileY:         0,
		windowCounter: 0,
		inWindow:      false,
		isFirstTile:   false,
	}
}

func (f *BgFetcher) reset(lineX byte, lineY byte) {
	f.state = ReadTileID

	if f.lcdc.windowEnabled == 1 && lineY >= f.scs.wy && lineX >= (f.scs.wx-7) {
		f.pixelY = f.windowCounter
		f.pixelX = lineX - (f.scs.wx - 7)
		f.inWindow = true
	} else {
		f.pixelY = f.scs.scy + lineY
		f.pixelX = f.scs.scx + lineX
		f.inWindow = false
	}
	f.tileY = (f.pixelY >> 3) & 31

	f.fetcherX = 0
	f.fifo.clear()
}

func (f *BgFetcher) cycle(shouldCycle bool) {
	if !shouldCycle {
		return
	}

	switch f.state {
	case ReadTileID:
		f.mapAddr = f.getTileMapBase() + uint16(f.tileY)*32 + uint16(((f.pixelX>>3)+f.fetcherX)&31)
		f.tileId = f.bus.PpuReadVram(f.mapAddr)
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
				if err := f.fifo.push(newPixel(f.tileData[i], 0xFF47, 0)); err != nil {
					panic(err)
				}
			}
			f.fetcherX++
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

func (f *BgFetcher) resetIfWindow(lineX byte, lineY byte) {
	// stop pushing to FIFO, we reached the window
	realWx := f.scs.wx - 7
	if !f.inWindow && f.lcdc.windowEnabled == 1 && lineY >= f.scs.wy && lineX >= realWx {
		f.reset(lineX, lineY)
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
