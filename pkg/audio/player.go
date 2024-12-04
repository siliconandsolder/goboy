package audio

// typedef unsigned char Uint8;
// void Callback(void *player, Uint8 *stream, int len);
import "C"
import (
	"github.com/veandco/go-sdl2/sdl"
	"reflect"
	"runtime"
	"unsafe"
)

const SAMPLE_RATE = 87
const AUDIO_FREQ = 48000

type Player struct {
	channel     chan byte
	numChannels int
	pinner      *runtime.Pinner
}

func NewPlayer() *Player {
	return &Player{
		channel:     make(chan byte, AUDIO_FREQ),
		numChannels: 0,
		pinner:      new(runtime.Pinner),
	}
}

func (p *Player) Start() error {
	if err := sdl.Init(sdl.INIT_AUDIO); err != nil {
		return err
	}

	p.pinner.Pin(&p.channel)

	spec := &sdl.AudioSpec{
		Freq:     AUDIO_FREQ,
		Format:   sdl.AUDIO_U8,
		Channels: 2,
		Samples:  AUDIO_FREQ / 60,
		Callback: sdl.AudioCallback(C.Callback),
		UserData: unsafe.Pointer(&p.channel),
	}

	if err := sdl.OpenAudio(spec, nil); err != nil {
		return err
	}

	sdl.PauseAudio(false)
	return nil
}

func (p *Player) Close() {
	sdl.PauseAudio(true)
	sdl.CloseAudio()
	p.pinner.Unpin()
}

func (p *Player) SendSample(sample byte) {
	p.channel <- sample
}

//export Callback
func Callback(player unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length)
	hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(stream)), Len: n, Cap: n}
	buf := *(*[]C.Uint8)(unsafe.Pointer(&hdr))
	channel := *(*chan byte)(player)

	for i := 0; i < n; i += 2 {
		var output C.Uint8 = 0
		select {
		case sample := <-channel:
			output = (C.Uint8)(sample)
		default:
			output = 0
		}
		buf[i] = output
		buf[i+1] = output
	}
}
