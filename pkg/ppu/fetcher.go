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
	shouldCycle bool
	state       FetcherState
	tileData    []byte

	tileIdx  byte
	tileId   byte
	mapAddr  uint16
	tileLine byte
}

func newFetcher(bus *bus.Bus) *Fetcher {
	return &Fetcher{
		fifo:        newFIFO(),
		bus:         bus,
		shouldCycle: false,
		state:       0,
		tileData:    make([]byte, 8),
		tileIdx:     0,
		tileId:      0,
		mapAddr:     0,
		tileLine:    0,
	}
}

func (f *Fetcher) reset(mapAddr uint16, tileLine byte) {
	f.tileIdx = 0
	f.mapAddr = mapAddr
	f.tileLine = tileLine
	f.state = ReadTileID

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
		f.tileId = f.bus.Read(f.mapAddr + uint16(f.tileIdx))
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
			// TODO: check if tile is flipped horizontally
			for i := 7; i >= 0; i-- {
				if err := f.fifo.push(f.tileData[i]); err != nil {
					panic(err)
				}
			}

			f.tileIdx++
			f.state = ReadTileID
		}
		break
	}
}

func (f *Fetcher) readTileLine(isHigh bool) {
	offset := TILE_DATA_START + uint16(f.tileId)<<4
	addr := offset + uint16(f.tileLine)<<2

	if isHigh {
		addr++
	}
	data := f.bus.Read(addr)

	for bitPos := byte(0); bitPos <= 7; bitPos++ {
		if !isHigh {
			f.tileData[bitPos] = (data >> bitPos) & 1
		} else {
			f.tileData[bitPos] |= ((data >> bitPos) & 1) << 1
		}
	}
}
