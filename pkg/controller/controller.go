package controller

import (
	"github.com/veandco/go-sdl2/sdl"
	"os"
)

const (
	RIGHT = iota
	LEFT
	UP
	DOWN
	A_BUTTON
	B_BUTTON
	SELECT
	START
)

type Controller struct {
	inputs        []bool
	selectButtons bool
	selectDPad    bool
}

func NewController() *Controller {
	return &Controller{
		inputs:        make([]bool, 8),
		selectButtons: true,
		selectDPad:    true,
	}
}

func (c *Controller) CheckJoypad() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			os.Exit(0)
		case *sdl.KeyboardEvent:
			keyCode := t.Keysym.Sym
			if t.State == sdl.PRESSED {
				switch keyCode {
				case sdl.K_RIGHT:
					c.inputs[RIGHT] = true
				case sdl.K_LEFT:
					c.inputs[LEFT] = true
				case sdl.K_UP:
					c.inputs[UP] = true
				case sdl.K_DOWN:
					c.inputs[DOWN] = true
				case sdl.K_z:
					c.inputs[A_BUTTON] = true
				case sdl.K_x:
					c.inputs[B_BUTTON] = true
				case sdl.K_RSHIFT:
					c.inputs[SELECT] = true
				case sdl.K_RETURN:
					c.inputs[START] = true
				}
			} else {
				switch keyCode {
				case sdl.K_RIGHT:
					c.inputs[RIGHT] = false
				case sdl.K_LEFT:
					c.inputs[LEFT] = false
				case sdl.K_UP:
					c.inputs[UP] = false
				case sdl.K_DOWN:
					c.inputs[DOWN] = false
				case sdl.K_z:
					c.inputs[A_BUTTON] = false
				case sdl.K_x:
					c.inputs[B_BUTTON] = false
				case sdl.K_RSHIFT:
					c.inputs[SELECT] = false
				case sdl.K_RETURN:
					c.inputs[START] = false
				}
			}

		default:
			break
		}
	}

	for _, val := range c.inputs {
		if val {
			return true
		}
	}

	return false
}

func (c *Controller) SetButtonSelectors(val byte) {
	c.selectDPad = false
	c.selectButtons = false

	if (val>>4)&1 == 0 {
		c.selectDPad = true
	}

	if (val>>5)&1 == 0 {
		c.selectButtons = true
	}
}

func (c *Controller) GetJoypadValue() byte {
	if c.selectButtons && c.selectDPad {
		return 0x3F
	}

	var selectVals byte = 0
	var buttonVals byte = 0
	buttonSlice := make([]bool, 3)

	if c.selectDPad {
		selectVals = 0b11101111
		buttonSlice = c.inputs[RIGHT : DOWN+1]
	} else if c.selectButtons {
		selectVals = 0b11011111
		buttonSlice = c.inputs[A_BUTTON:]
	}

	for idx, val := range buttonSlice {
		if val {
			buttonVals |= 1 << idx
		}
	}

	buttons := ^buttonVals | 0xF0

	return selectVals & buttons
}
