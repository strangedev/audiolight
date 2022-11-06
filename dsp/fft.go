package dsp

import (
	"context"
	"github.com/mjibson/go-dsp/dsputils"
	"github.com/mjibson/go-dsp/fft"
	"github.com/strangedev/audiolight/audio"
	"math/cmplx"
)

type FFTFrame = []float64

func FFT(ctx context.Context, in <-chan audio.Frame, frameSize int) <-chan FFTFrame {
	out := make(chan FFTFrame, 16)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case frame := <-in:
				fftResult := fft.FFT(dsputils.ToComplex(frame))
				fftFrame := make(FFTFrame, frameSize/2)

				for i := range fftFrame {
					fftFrame[i] = cmplx.Abs(fftResult[i])
				}

				out <- fftFrame
			}
		}
	}()

	return out
}

type FFTInterpreter struct {
	audio.RecordingOptions
}

func NewFFTInterpreter(options audio.RecordingOptions) FFTInterpreter {
	return FFTInterpreter{options}
}

type FrequencyContent struct {
	Frequency float64
	Intensity float64
}

func (interpreter FFTInterpreter) GetFFTBinFrequency(binIndex int) float64 {
	return float64(binIndex) * interpreter.SampleRate / float64(interpreter.FrameSize)
}

func (interpreter FFTInterpreter) GetFrequencyContent(frame FFTFrame) []FrequencyContent {
	result := make([]FrequencyContent, len(frame))

	for i, value := range frame {
		result[i] = FrequencyContent{
			Frequency: float64(i) * interpreter.SampleRate / float64(interpreter.FrameSize),
			Intensity: value,
		}
	}

	return result
}

func (interpreter FFTInterpreter) GetBinCount() int {
	return interpreter.FrameSize / 2
}
