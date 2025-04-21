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
	"os"
)

const (
	defaultScale            = 4
	gbWidth, gbHeight int32 = 160, 144
	scaleFName              = "scale"
	romFName                = "rom"
)

var romName string
var scale int32

var rootCmd = &cobra.Command{
	Use:   "goboy",
	Short: "a gameboy emulator written in Golang",

	Run: func(cmd *cobra.Command, args []string) {
		fileName, _ := cmd.Flags().GetString(romFName)
		if fileName == "" {
			panic("no rom :(") // TODO: splash screen
		}

		scale, _ := cmd.Flags().GetInt32(scaleFName)
		var winWidth, winHeight = gbWidth * scale, gbHeight * scale

		fileData, err := os.ReadFile(fileName)
		if err != nil {
			panic(err) // no point in continuing
		}

		cart := cartridge.NewCartridge(fileData)

		var window *sdl.Window
		var renderer *sdl.Renderer
		var texture *sdl.Texture

		window, err = sdl.CreateWindow(fmt.Sprintf("GOBOY - %s", cart.Title), sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
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

		player := audio.NewPlayer()
		if err := player.Start(); err != nil {
			panic(err)
		}
		defer func(player *audio.Player) {
			player.Close()
		}(player)

		ctrl := controller.NewController()
		m := interrupts.NewManager()
		s := audio.NewSoundChip(player)
		b := bus.NewBus(cart, m, ctrl, s)
		t := cpu.NewSysTimer(b)
		c := cpu.NewCpu(b, m, t)
		p := ppu.NewPPU(b)

		cart.LoadRAMFromFile()
		defer cart.SaveRAMToFile()

		var vBuffer []uint32

		running := true
		for running {
			cycles, err := c.Cycle()
			if err != nil {
				panic(err)
			}
			t.Cycle(cycles)
			s.Cycle(cycles)
			cart.UpdateCounter(cycles)

			if vBuffer, err = p.Cycle(cycles); err != nil {
				panic(err)
			} else if vBuffer != nil {
				pixels, _, err := texture.Lock(nil)
				if err != nil {
					panic(err)
				}

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

				if err := renderer.Clear(); err != nil {
					panic(err)
				}
				if err := renderer.Copy(texture, nil, nil); err != nil {
					panic(err)
				}
				renderer.Present()

				//if ctrl.CheckJoypad() {
				//	b.ToggleInterrupt(interrupts.JOYPAD)
				//}

				for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
					switch t := event.(type) {
					case *sdl.KeyboardEvent:
						keyCode := t.Keysym.Sym
						if keyCode == sdl.K_ESCAPE {
							running = false
						} else {
							ctrl.CheckJoypadEvent(keyCode, t.State)
						}
					case *sdl.QuitEvent:
						running = false
					default:
						break
					}
				}
			}
		}
	},
}

func main() {
	rootCmd.Flags().Int32Var(&scale, scaleFName, defaultScale, "scale the window size as a multiple of the default gameboy resolution")
	rootCmd.Flags().StringVar(&romName, romFName, "", "specify a .gb file")
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
