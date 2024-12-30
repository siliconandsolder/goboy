package audio

// typedef unsigned char Uint8;
// void Callback(void *queue, Uint8 *stream, int len);
import "C"
import (
	"github.com/veandco/go-sdl2/sdl"
	"reflect"
	"runtime"
	"unsafe"
)

const AUDIO_FREQUENCY = 48000

type stereoSample struct {
	leftSample  byte
	rightSample byte
}

type Player struct {
	channel     chan stereoSample
	numChannels int
	pinner      *runtime.Pinner
}

func NewPlayer() *Player {
	return &Player{
		channel:     make(chan stereoSample, AUDIO_FREQUENCY),
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
		Freq:     AUDIO_FREQUENCY,
		Format:   sdl.AUDIO_U8,
		Channels: 2,
		Samples:  AUDIO_FREQUENCY / 60,
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

func (p *Player) SendSample(sample stereoSample) {
	p.channel <- sample
}

//export Callback
func Callback(queue unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length)
	hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(stream)), Len: n, Cap: n}
	buf := *(*[]C.Uint8)(unsafe.Pointer(&hdr))
	channel := *(*chan stereoSample)(queue)

	for i := 0; i < n; i += 2 {
		var output stereoSample
		select {
		case sample := <-channel:
			output = sample
		default:
			output = stereoSample{}
		}
		buf[i] = (C.Uint8)(output.leftSample)
		buf[i+1] = (C.Uint8)(output.rightSample)
	}
}
