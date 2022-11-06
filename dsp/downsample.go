package dsp

import (
	"context"
	"github.com/strangedev/audiolight/audio"
	"github.com/strangedev/audiolight/math"
)

// Downsample downsamples an input stream by averaging every N frames.
func Downsample[TSample math.Number](ctx context.Context, in <-chan []TSample, frameSize int, reductionRate int) <-chan []TSample {
	out := make(chan []TSample, audio.ChannelBufferSize)

	go func() {
		defer close(out)

		accumulatorFrame := make([]TSample, frameSize)
		outFrame := make([]TSample, frameSize)
		inFrameCounter := 0

		for {
			select {
			case <-ctx.Done():
				return
			case frame := <-in:
				for i, sample := range frame {
					currentAverage := accumulatorFrame[i]
					newAverage := math.AddToAverage(currentAverage, sample, TSample(inFrameCounter))
					accumulatorFrame[i] = newAverage
				}

				inFrameCounter++

				if inFrameCounter == reductionRate {
					copy(outFrame, accumulatorFrame)
					out <- outFrame

					for i := range accumulatorFrame {
						accumulatorFrame[i] = 0
					}
					inFrameCounter = 0
				}
			}
		}
	}()

	return out
}
