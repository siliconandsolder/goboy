package ppu

import "github.com/siliconandsolder/go-boy/pkg/bus"

const (
	SPRITE_DATA_ZERO = 0xFF48
	SPRITE_DATA_ONE  = 0xFF49

	SPRITE_TILE_ADDR = 0x8000
)

type SpriteFetcher struct {
	spriteToFetch *OamObj
	fifo          *PixelFIFO
	state         FetcherState
	bus           *bus.Bus
	lcdc          *LcdControl
	dataLow       byte
	dataHigh      byte
	tileData      []byte
	lineY         byte
}

func newSpriteFetcher(bus *bus.Bus, lcdc *LcdControl) *SpriteFetcher {
	return &SpriteFetcher{
		spriteToFetch: nil,
		fifo:          newFIFO(false),
		state:         ReadTileID,
		bus:           bus,
		lcdc:          lcdc,
		dataLow:       0,
		dataHigh:      0,
		tileData:      make([]byte, 8),
		lineY:         0,
	}
}

func (s *SpriteFetcher) reset(lineY byte) {
	s.state = ReadTileID
	s.lineY = lineY
	s.fifo.clear()
}

func (s *SpriteFetcher) cycle(shouldCycle bool) {
	if !shouldCycle {
		return
	}

	switch s.state {
	case ReadTileID:
		s.state = ReadTileData0
		break
	case ReadTileData0:
		s.readTileData(false)
		s.state = ReadTileData1
		break
	case ReadTileData1:
		s.readTileData(true)
		s.state = PushToFIFO
		break
	case PushToFIFO:
		tempFifo := newFIFO(false)

		for i := byte(0); i <= 7; i++ {
			pixelBit := i
			if s.spriteToFetch.attributes.xFlip == 0 {
				pixelBit = 7 - pixelBit
			}

			var colourNum byte = 0
			colourNum = ((s.dataHigh >> pixelBit) & 1) << 1
			colourNum |= (s.dataLow >> pixelBit) & 1

			palette := s.spriteToFetch.attributes.palette
			priority := s.spriteToFetch.attributes.priority
			var paletteAddr uint16
			if palette == 0 {
				paletteAddr = SPRITE_DATA_ZERO
			} else {
				paletteAddr = SPRITE_DATA_ONE
			}

			if err := tempFifo.push(newPixel(colourNum, paletteAddr, priority)); err != nil {
				panic(err)
			}
		}

		s.mixFifos(tempFifo)
		s.spriteToFetch = nil
		s.state = ReadTileID
		break
	}
}

func (s *SpriteFetcher) readTileData(isHigh bool) {
	tileNum := s.spriteToFetch.tileNum
	if s.lcdc.objSize == 1 {
		tileNum &= 0b11111110
	}
	tileAddr := SPRITE_TILE_ADDR + uint16(tileNum)*16
	if isHigh {
		tileAddr++
	}

	var spriteHeight uint16
	if s.lcdc.objSize == 0 {
		spriteHeight = 8
	} else {
		spriteHeight = 16
	}

	offset := ((uint16(s.lineY) - uint16(s.spriteToFetch.posY-16)) % spriteHeight) * 2
	if s.spriteToFetch.attributes.yFlip == 1 {
		offset = (spriteHeight-1)*2 - offset
	}
	data := s.bus.PpuReadVram(tileAddr + offset)

	if isHigh {
		s.dataHigh = data
	} else {
		s.dataLow = data
	}
}

// hacky bullshit
func (s *SpriteFetcher) mixFifos(tempFifo *PixelFIFO) {
	for i := 0; i < MAX_SIZE_FG; i++ {
		if s.fifo.queue[i] == nil {
			if err := s.fifo.push(tempFifo.queue[i]); err != nil {
				panic(err)
			}
			continue
		}
		if tempFifo.queue[i].colourNum != 0 && s.fifo.queue[i].colourNum == 0 {
			s.fifo.queue[i].colourNum = tempFifo.queue[i].colourNum
			s.fifo.queue[i].priority = tempFifo.queue[i].priority
			s.fifo.queue[i].paletteAddr = tempFifo.queue[i].paletteAddr
		}
	}
}
