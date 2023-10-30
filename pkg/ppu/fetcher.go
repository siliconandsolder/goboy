package ppu

import "github.com/siliconandsolder/go-boy/pkg/bus"

type FetcherState byte

const (
	ReadTileID FetcherState = iota
	ReadTileData0
	ReadTileData1
	PushToFIFO
)

type Fetcher struct {
	fifo        *PixelFIFO
	bus         *bus.Bus
	lcdc        *LcdControl
	scs         *ScrollStatus
	shouldCycle bool
	state       FetcherState
	tileData    []byte

	tileIdx    byte
	tileId     byte
	mapAddr    uint16
	tileLine   byte
	tileOffset int16
	lineX      byte
	pixelX     byte
	tileX      byte
	lineY      byte
	pixelY     byte
	tileY      byte
}

func newFetcher(bus *bus.Bus, lcdc *LcdControl, scs *ScrollStatus) *Fetcher {
	return &Fetcher{
		fifo:        newFIFO(),
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
	}
}

func (f *Fetcher) reset(lineY byte) {
	f.lineX = 0
	f.lineY = lineY
	f.state = ReadTileID

	if f.lcdc.windowEnabled == 1 && f.lineY >= f.scs.wy {
		f.pixelY = lineY - f.scs.wy
	} else {
		f.pixelY = f.scs.scy + lineY
	}
	f.tileY = (f.pixelY >> 3) & 31

	f.calculateTileMapAddr()
	f.fifo.clear()
}

func (f *Fetcher) cycle() {
	if !f.shouldCycle {
		f.shouldCycle = true
		return
	}

	f.shouldCycle = false

	switch f.state {
	case ReadTileID:
		// TODO: something to do with the tileline. Calculating X every cycle might be causing repeat tile
		//f.calculateTileMapAddr()
		f.mapAddr = f.getTileMapBase() + uint16(f.tileY)*32 + uint16(f.tileX)
		f.tileId = f.bus.PpuReadVram(f.mapAddr)
		if f.lcdc.tileDataArea == 1 {
			f.tileOffset = int16(f.tileId)
		} else {
			f.tileOffset = int16(int8(f.tileId)) + 128
		}
		f.tileOffset *= 16
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

func (f *Fetcher) readTileLine(isHigh bool) {
	// get tile data base
	addr := f.getTileDataBase() + uint16(f.tileOffset) + uint16(f.pixelY%8)<<1

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
}

func (f *Fetcher) calculateTileMapAddr() {
	realWx := f.scs.wx - 7
	if f.lcdc.windowEnabled == 1 && f.lineY >= f.scs.wy && f.lineX >= realWx {
		f.pixelX = f.lineX - realWx
	} else {
		f.pixelX = f.scs.scx + f.lineX
	}

	f.tileX = (f.pixelX >> 3) & 31

	var tileMapBase uint16 = 0

	if f.lcdc.windowEnabled == 1 && f.lineY >= f.scs.wy && f.lineX >= realWx {
		if f.lcdc.wTileMapArea == 1 {
			tileMapBase = TILE_MAP_START_ONE
		} else {
			tileMapBase = TILE_MAP_START_ZERO
		}
	} else if f.lcdc.bgTileMapArea == 1 {
		tileMapBase = TILE_MAP_START_ONE
	}

	tileMapBase = TILE_MAP_START_ZERO

	f.mapAddr = tileMapBase + uint16(f.tileY)*32 + uint16(f.tileX)
}

func (f *Fetcher) resetIfWindow() {
	// stop pushing to FIFO, we reached the window
	realWx := f.scs.wx - 7
	if f.lcdc.windowEnabled == 1 && f.lcdc.wTileMapArea == 1 && f.lineY >= f.scs.wy && f.lineX >= realWx {
		f.reset(f.lineY)
	}
}

func (f *Fetcher) getTileDataBase() uint16 {
	if f.lcdc.tileDataArea == 1 {
		return TILE_DATA_START_ZERO
	} else {
		return TILE_DATA_START_ONE
	}
}

func (f *Fetcher) getTileMapBase() uint16 {
	realWx := f.scs.wx - 7
	if f.lcdc.windowEnabled == 1 && f.lineY >= f.scs.wy && f.lineX >= realWx {
		if f.lcdc.wTileMapArea == 1 {
			return TILE_MAP_START_ONE
		} else {
			return TILE_MAP_START_ZERO
		}
	} else if f.lcdc.bgTileMapArea == 1 {
		return TILE_MAP_START_ONE
	}

	return TILE_MAP_START_ZERO
}
