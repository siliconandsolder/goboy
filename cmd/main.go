package main

import (
	"image"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/siliconandsolder/go-boy/pkg/audio"
	"github.com/siliconandsolder/go-boy/pkg/bus"
	"github.com/siliconandsolder/go-boy/pkg/cartridge"
	"github.com/siliconandsolder/go-boy/pkg/controller"
	"github.com/siliconandsolder/go-boy/pkg/cpu"
	"github.com/siliconandsolder/go-boy/pkg/interrupts"
	"github.com/siliconandsolder/go-boy/pkg/ppu"
	"github.com/spf13/cobra"
)

const (
	defaultScale             = 4
	gbWidth, gbHeight uint32 = 160, 144
	scaleFName               = "scale"
	romFName                 = "rom"
)

var romName string
var scale uint32

var rootCmd = &cobra.Command{
	Use:   "goboy",
	Short: "a gameboy emulator written in Golang",

	Run: func(cmd *cobra.Command, args []string) {
		fileName, _ := cmd.Flags().GetString(romFName)
		if fileName == "" {
			panic("no rom :(") // TODO: splash screen
		}

		fileData, err := os.ReadFile(fileName)
		if err != nil {
			panic(err) // no point in continuing
		}

		cart := cartridge.NewCartridge(fileData)

		var winWidth, winHeight = float32(gbWidth * scale), float32(gbHeight * scale)

		myApp := app.New()
		myWindow := myApp.NewWindow("GOBOY - " + cart.Title)

		texture := image.NewRGBA(image.Rect(0, 0, int(gbWidth), int(gbHeight)))
		texCanvas := canvas.NewImageFromImage(texture)
		texCanvas.Image = texture
		texCanvas.ScaleMode = canvas.ImageScalePixels
		texCanvas.Resize(fyne.NewSize(winWidth, winHeight))
		texCanvas.SetMinSize(fyne.NewSize(winWidth, winHeight))

		myWindow.SetContent(texCanvas)
		myWindow.Resize(texCanvas.Size())
		myWindow.CenterOnScreen()
		myWindow.SetFixedSize(true)

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

		running := true
		paused := false

		var deskCanvas desktop.Canvas
		if desk, ok := myWindow.Canvas().(desktop.Canvas); ok {
			deskCanvas = desk
		} else {
			panic("Canvas does not implement desktop.Canvas")
		}

		deskCanvas.SetOnKeyUp(func(event *fyne.KeyEvent) {
			switch event.Name {
			case fyne.KeyRight:
				ctrl.SetKeyUp(controller.RIGHT)
			case fyne.KeyLeft:
				ctrl.SetKeyUp(controller.LEFT)
			case fyne.KeyUp:
				ctrl.SetKeyUp(controller.UP)
			case fyne.KeyDown:
				ctrl.SetKeyUp(controller.DOWN)
			case fyne.KeyZ:
				ctrl.SetKeyUp(controller.A_BUTTON)
			case fyne.KeyX:
				ctrl.SetKeyUp(controller.B_BUTTON)
			case fyne.KeySpace:
				ctrl.SetKeyUp(controller.SELECT)
			case fyne.KeyReturn:
				ctrl.SetKeyUp(controller.START)
			}
		})
		deskCanvas.SetOnKeyDown(func(event *fyne.KeyEvent) {
			if event.Name == fyne.KeyEscape {
				running = false
				myApp.Quit()
			}

			switch event.Name {
			case fyne.KeyRight:
				ctrl.SetKeyDown(controller.RIGHT)
			case fyne.KeyLeft:
				ctrl.SetKeyDown(controller.LEFT)
			case fyne.KeyUp:
				ctrl.SetKeyDown(controller.UP)
			case fyne.KeyDown:
				ctrl.SetKeyDown(controller.DOWN)
			case fyne.KeyZ:
				ctrl.SetKeyDown(controller.A_BUTTON)
			case fyne.KeyX:
				ctrl.SetKeyDown(controller.B_BUTTON)
			case fyne.KeySpace:
				ctrl.SetKeyDown(controller.SELECT)
			case fyne.KeyReturn:
				ctrl.SetKeyDown(controller.START)
			}
		})

		go func() {
			for running && !paused {
				cycles, err := c.Cycle()
				if err != nil {
					panic(err)
				}
				t.Cycle(cycles)
				s.Cycle(cycles)
				cart.UpdateCounter(cycles)

				if vBuffer, err := p.Cycle(cycles); err != nil {
					panic(err)
				} else if vBuffer != nil {
					for y, scanLine := range vBuffer {
						for x, pixel := range scanLine {
							red := uint8(pixel >> 24)
							green := uint8(pixel >> 16 & 0xFF)
							blue := uint8(pixel >> 8 & 0xFF)
							texture.Set(x, y, color.RGBA{R: red, G: green, B: blue, A: 0xff})
						}
					}

					fyne.Do(func() {
						texCanvas.Refresh()
					})

					if ctrl.CheckForInputs() {
						b.ToggleInterrupt(interrupts.JOYPAD)
					}
				}
			}
		}()

		myWindow.ShowAndRun()
	},
}

func main() {
	rootCmd.Flags().Uint32Var(&scale, scaleFName, defaultScale, "scale the window size as a multiple of the default gameboy resolution")
	rootCmd.Flags().StringVar(&romName, romFName, "", "specify a .gb file")
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
