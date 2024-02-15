package ppu

import "fmt"

const MAX_SIZE_BG = 16
const MAX_SIZE_FG = 8

type Pixel struct {
	colourNum   byte
	paletteAddr uint16
	priority    byte
}

func newPixel(colourNum byte, paletteAddr uint16, priority byte) *Pixel {
	return &Pixel{
		colourNum:   colourNum,
		paletteAddr: paletteAddr,
		priority:    priority,
	}
}

type PixelFIFO struct {
	queue        []*Pixel
	size         int
	isBackground bool
}

func newFIFO(isBackground bool) *PixelFIFO {
	var queue []*Pixel
	if isBackground {
		queue = make([]*Pixel, MAX_SIZE_BG)
	} else {
		queue = make([]*Pixel, MAX_SIZE_FG)
	}
	return &PixelFIFO{
		queue:        queue,
		size:         0,
		isBackground: isBackground,
	}
}

func (p *PixelFIFO) push(pixel *Pixel) error {
	if p.size == MAX_SIZE_BG {
		return fmt.Errorf("pixel fifo is full")
	}
	p.size++

	p.queue[p.size-1] = pixel
	return nil
}

func (p *PixelFIFO) pop() *Pixel {
	if p.size == 0 {
		return nil
	}
	p.size--

	retVal := p.queue[0]

	for i := 0; i < p.size; i++ {
		p.queue[i] = p.queue[i+1]
	}
	p.queue[p.size] = nil // zero the back of the queue

	return retVal
}

func (p *PixelFIFO) clear() {
	p.size = 0
	if p.isBackground {
		p.queue = make([]*Pixel, MAX_SIZE_BG)
	} else {
		p.queue = make([]*Pixel, MAX_SIZE_FG)
	}
}

func (p *PixelFIFO) isEmpty() bool {
	return p.size == 0
}
