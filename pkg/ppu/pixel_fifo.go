package ppu

import "fmt"

const MAX_SIZE = 16

type PixelFIFO struct {
	queue []byte
	size  int
}

func newFIFO() *PixelFIFO {
	return &PixelFIFO{
		queue: make([]byte, 16),
		size:  0,
	}
}

func (p *PixelFIFO) push(pixel byte) error {
	if p.size == MAX_SIZE {
		return fmt.Errorf("pixel fifo is full")
	}
	p.size++

	p.queue[p.size-1] = pixel
	return nil
}

func (p *PixelFIFO) pop() (byte, bool) {
	if p.size == 0 {
		return 0, false
	}
	p.size--

	retVal := p.queue[0]

	for i := 0; i < p.size; i++ {
		p.queue[i] = p.queue[i+1]
	}
	p.queue[p.size] = 0 // zero the back of the queue

	return retVal, true
}

func (p *PixelFIFO) clear() {
	p.size = 0
	p.queue = make([]byte, 16)
}

func (p *PixelFIFO) isEmpty() bool {
	return p.size == 0
}
