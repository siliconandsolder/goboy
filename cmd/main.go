package main

import (
	"fmt"
	"github.com/siliconandsolder/go-boy/pkg/bus"
	"github.com/siliconandsolder/go-boy/pkg/cartridge"
	"github.com/siliconandsolder/go-boy/pkg/cpu"
	"github.com/siliconandsolder/go-boy/pkg/interrupts"
	"github.com/siliconandsolder/go-boy/pkg/ppu"
	"github.com/spf13/cobra"
	"github.com/veandco/go-sdl2/sdl"
)

var cmd = &cobra.Command{
	Use:   "test",
	Short: "my emulator",

	Run: func(cmd *cobra.Command, args []string) {
		var winWidth, winHeight int32 = 640, 576
		var gbWidth, gbHeight int32 = 160, 144
		var window *sdl.Window
		var renderer *sdl.Renderer
		var texture *sdl.Texture
		var err error

		window, err = sdl.CreateWindow("GOBOY", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
			winWidth, winHeight, sdl.WINDOW_SHOWN)
		defer window.Destroy()

		renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
		if err != nil {
			panic(fmt.Sprintf("Failed to create renderer: %s\n", err))
		}
		defer renderer.Destroy()

		texture, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_STREAMING, gbWidth, gbHeight)
		if err != nil {
			panic(err)
		}
		defer texture.Destroy()

		//rect := sdl.Rect{
		//	X: 0,
		//	Y: 0,
		//	W: winWidth,
		//	H: winHeight,
		//}

		cart := cartridge.NewCartridge("./roms/mts/acceptance/oam_dma/basic.gb")
		m := interrupts.NewManager()
		b := bus.NewBus(cart, m)
		t := cpu.NewSysTimer(b)
		c := cpu.NewCpu(b, m, t)
		p := ppu.NewPPU(b)

		// TODO: return cycles from cpu, pass them to ppu and then timer

		for {
			cycles, err := c.Cycle()
			if err != nil {
				panic(err)
			}
			t.Cycle(cycles)
			if buffer, err := p.Cycle(cycles); err != nil {
				panic(err)
			} else if buffer != nil {
				pixels, _, err := texture.Lock(nil)
				//fmt.Sprintf("pixels: %v\n", pixels)
				for i := 0; i < len(buffer); i++ {
					red := byte(buffer[i] >> 24)
					green := byte(buffer[i] >> 16 & 0xFF)
					blue := byte(buffer[i] >> 8 & 0xFF)

					pixels[i*4] = 0xFF // alpha
					pixels[i*4+1] = blue
					pixels[i*4+2] = green
					pixels[i*4+3] = red
				}
				if err != nil {
					panic(err)
				}
				texture.Unlock()

				renderer.Clear()
				renderer.Copy(texture, nil, nil)
				renderer.Present()
			}
		}
	},
}

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
