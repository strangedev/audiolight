package audio

import (
	"context"
	"errors"
	"fmt"
	"github.com/gordonklaus/portaudio"
)

const FrameSize = 64

type Frame = []float32

type RecordingOptions struct {
	SampleRate float64
}

func RecordAudio(ctx context.Context, options RecordingOptions) (<-chan Frame, error) {
	frameChan := make(chan Frame, 16)

	if err := portaudio.Initialize(); err != nil {
		close(frameChan)
		return frameChan, err
	}

	go func() {
		defer func() {
			close(frameChan)

			if err := portaudio.Terminate(); err != nil {
				panic(err)
			}
		}()

		handleAudioInput := func(in [][]float32, timeInfo portaudio.StreamCallbackTimeInfo, flags portaudio.StreamCallbackFlags) {
			if len(in) != 1 {
				panic(errors.New(fmt.Sprintf("received an unexpected number of input channels, expected %d but got %d", 1, len(in))))
			}

			frame := in[0]
			frameChan <- frame
		}

		stream, err := portaudio.OpenDefaultStream(
			1,
			0,
			options.SampleRate,
			FrameSize,
			handleAudioInput,
		)
		if err != nil {
			// TODO: send error frame via frameChan
			return
		}

		if err := stream.Start(); err != nil {
			return // TODO: send error frame via frameChan
		}
		defer func() {
			if err := stream.Close(); err != nil {
				panic(err)
			}
		}()

		<-ctx.Done()
	}()

	return frameChan, nil
}
