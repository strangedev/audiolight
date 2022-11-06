package dsp

import (
	"context"
	"github.com/strangedev/audiolight/math"
)

// DownsampleDynamically downsamples an input stream by averaging frames. The reduction rate adjust dynamically with
// the back pressure exerted on the output stream. When the output stream is blocked, frames are collected and
// averaged. When the output stream becomes unblocked, the downsampled frame is sent.
// This is useful if a source stream is producing more frames than can be handled downstream.
func DownsampleDynamically[TSample math.Number](ctx context.Context, in <-chan []TSample, frameSize int) <-chan []TSample {
	out := make(chan []TSample, 1)

	go func() {
		defer close(out)

		outFrame := make([]TSample, frameSize)
		inFrameCounter := TSample(0)
		resetOutFrame := func() {
			for i := range outFrame {
				outFrame[i] = 0
			}
			inFrameCounter = 0
		}

	Loop:
		for {
			select {
			case <-ctx.Done():
				return
			case frame := <-in:
				for i, sample := range frame {
					currentAverage := outFrame[i]

					outFrame[i] = math.AddToAverage(currentAverage, sample, inFrameCounter)
				}

				inFrameCounter++

				select {
				case out <- outFrame:
					resetOutFrame()
				default:
					continue Loop
				}
			}
		}
	}()

	return out
}
