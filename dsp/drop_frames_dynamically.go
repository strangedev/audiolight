package dsp

import (
	"context"
	"github.com/strangedev/audiolight/math"
)

// DropFramesDynamically matches the flow rate of an input stream to the capacity of the output stream by dropping frames.
// When the output stream is blocked, frames are dropped from the input stream. When the output stream becomes unblocked,
// the most recently received input frame is sent to the output stream.
func DropFramesDynamically[TSample math.Number](ctx context.Context, in <-chan []TSample) <-chan []TSample {
	out := make(chan []TSample, 1)

	go func() {
		defer close(out)

		lastFrameChan := make(chan []TSample, 1)

		for {
			select {
			case <-ctx.Done():
				return
			case lastFrame := <-lastFrameChan:
				// when there's a frame stored
				select {
				case out <- lastFrame:
					// and we can send it, send it
					continue
				case newFrame := <-in:
					// and we cannot send it, but there's already a new frame, retry with the new frame
					lastFrameChan <- newFrame
				default:
					// and there's no new frame yet, retry sending the old frame
					lastFrameChan <- lastFrame
				}
			default:
				// when there's no frame, wait for one
				newFrame := <-in

				lastFrameChan <- newFrame
			}
		}
	}()

	return out
}
