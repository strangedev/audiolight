package audio

import (
	"context"
	"github.com/mjibson/go-dsp/dsputils"
	"github.com/mjibson/go-dsp/fft"
	"math/cmplx"
)

type FFTFrame = []complex128

func FFT(ctx context.Context, in <-chan Frame) (<-chan FFTFrame, error) {
	out := make(chan FFTFrame, 16)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case frame := <-in:
				fftResult := fft.FFT(dsputils.ToComplex(frame))

				out <- fftResult
			}
		}
	}()

	return out, nil
}

type FFTInterpreter struct {
	sampleRate float64
}

func NewFFTInterpreter(options RecordingOptions) FFTInterpreter {
	return FFTInterpreter{
		sampleRate: options.SampleRate,
	}
}

type FrequencyContent struct {
	Frequency float64
	Intensity float64
}

func (interpreter FFTInterpreter) GetFrequencyContent(frame FFTFrame) []FrequencyContent {
	result := make([]FrequencyContent, len(frame))

	for i, value := range frame {
		result[i] = FrequencyContent{
			Frequency: float64(i) * interpreter.sampleRate / FrameSize,
			Intensity: cmplx.Abs(value),
		}
	}

	return result
}
