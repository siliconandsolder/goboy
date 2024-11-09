package main

import (
	"fmt"
	"github.com/siliconandsolder/go-boy/pkg/audio"
	"github.com/siliconandsolder/go-boy/pkg/bus"
	"github.com/siliconandsolder/go-boy/pkg/cartridge"
	"github.com/siliconandsolder/go-boy/pkg/controller"
	"github.com/siliconandsolder/go-boy/pkg/cpu"
	"github.com/siliconandsolder/go-boy/pkg/interrupts"
	"github.com/siliconandsolder/go-boy/pkg/ppu"
	"github.com/spf13/cobra"
	"github.com/veandco/go-sdl2/sdl"
	"math"
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

		player := audio.NewPlayer()
		if err := player.Start(); err != nil {
			panic(err)
		}
		defer func(player *audio.Player) {
			err := player.Close()
			if err != nil {
				panic(err)
			}
		}(player)

		ctrl := controller.NewController()
		cart := cartridge.NewCartridge("./roms/mario.gb")
		m := interrupts.NewManager()
		s := audio.NewSoundChip(player)
		b := bus.NewBus(cart, m, ctrl, s)
		t := cpu.NewSysTimer(b)
		c := cpu.NewCpu(b, m, t)
		p := ppu.NewPPU(b)

		var end uint64 = 0
		start := sdl.GetPerformanceCounter()

		var vBuffer []uint32
		var prevDivTimer byte = 0

		for {
			cycles, err := c.Cycle()
			if err != nil {
				panic(err)
			}
			t.Cycle(cycles)
			s.Cycle(cycles)

			curDivTimer := b.Read(cpu.DIV_TIMER_ADDRESS)
			// check for falling edge on bit 5
			if (prevDivTimer>>5&1) == 1 && (curDivTimer>>5&1) == 0 {
				s.CycleFrameSequencer()
			}

			prevDivTimer = curDivTimer

			if vBuffer, err = p.Cycle(cycles); err != nil {
				panic(err)
			} else if vBuffer != nil {
				pixels, _, err := texture.Lock(nil)
				if err != nil {
					panic(err)
				}
				//fmt.Sprintf("pixels: %v\n", pixels)
				for i := 0; i < len(vBuffer); i++ {
					red := byte(vBuffer[i] >> 24)
					green := byte(vBuffer[i] >> 16 & 0xFF)
					blue := byte(vBuffer[i] >> 8 & 0xFF)

					pixels[i*4] = 0xFF // alpha
					pixels[i*4+1] = blue
					pixels[i*4+2] = green
					pixels[i*4+3] = red
				}

				texture.Unlock()

				renderer.Clear()
				renderer.Copy(texture, nil, nil)
				renderer.Present()

				if ctrl.CheckJoypad() {
					b.ToggleInterrupt(interrupts.JOYPAD)
				}

				end = sdl.GetPerformanceCounter()

				elapsedMS := float32(end-start) / float32(sdl.GetPerformanceFrequency()) * 1000.0
				delayMS := 16.666 - elapsedMS
				if delayMS >= 0.0 {
					sdl.Delay(uint32(math.Floor(float64(delayMS))))
				}

				start = sdl.GetPerformanceCounter()
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
