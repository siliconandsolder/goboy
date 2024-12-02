package audio

import "github.com/gordonklaus/portaudio"

const SAMPLE_RATE = 87
const AUDIO_FREQ = 4800

type Player struct {
	stream      *portaudio.Stream
	channel     chan byte
	numChannels int
}

func NewPlayer() *Player {
	return &Player{
		stream:      nil,
		channel:     make(chan byte, AUDIO_FREQ),
		numChannels: 0,
	}
}

func (p *Player) Start() error {
	if err := portaudio.Initialize(); err != nil {
		return err
	}

	host, err := portaudio.DefaultHostApi()
	if err != nil {
		return err
	}
	parameters := portaudio.HighLatencyParameters(nil, host.DefaultOutputDevice)
	parameters.Output.Channels = 2 // stereo
	parameters.SampleRate = float64(AUDIO_FREQ)
	stream, err := portaudio.OpenStream(parameters, p.Callback)

	if err != nil {
		return err
	}
	if err := stream.Start(); err != nil {
		return err
	}

	p.stream = stream
	p.numChannels = parameters.Output.Channels

	return nil
}

func (p *Player) Close() error {
	return p.stream.Close()
}

func (p *Player) SendSample(sample byte) {
	p.channel <- sample
}

// TODO: switch back to SDL
func (p *Player) Callback(buffer []byte) {
	var output byte

	for i := range buffer {
		if i%p.numChannels == 0 {
			select {
			case sample := <-p.channel:
				output = sample
			default:
				output = 0
			}
		}
		buffer[i] = output
	}
}
