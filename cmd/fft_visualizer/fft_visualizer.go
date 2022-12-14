package main

import (
	"context"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/strangedev/audiolight/audio"
	"github.com/strangedev/audiolight/dsp"
	"golang.org/x/image/colornames"
	"math"
	"os"
	"os/signal"
	"time"
)

const w, h = float64(200), float64(200)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, w, h),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	imd := imdraw.New(nil)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	recordingOptions := audio.RecordingOptions{
		SampleRate: 44100,
		FrameSize:  256,
	}

	audioFrames, err := audio.RecordAudio(ctx, recordingOptions)
	if err != nil {
		panic(err)
	}

	fftFrames := dsp.FFT(ctx, audioFrames, recordingOptions.FrameSize)
	downsampledFFTFrames := dsp.Downsample(ctx, fftFrames, recordingOptions.FrameSize, 4)
	pressureMatchedFFTFrames := dsp.DropFramesDynamically(ctx, downsampledFFTFrames)

	fftInterpreter := dsp.NewFFTInterpreter(recordingOptions)

	ticker := time.NewTicker(20 * time.Millisecond)

	for !win.Closed() {
		<-ticker.C

		imd.Clear()
		win.Clear(colornames.Aliceblue)

		frame := <-pressureMatchedFFTFrames

		for _, frequencyContent := range fftInterpreter.GetFrequencyContent(frame) {
			imd.Color = colornames.Limegreen

			x := w * frequencyContent.Frequency / recordingOptions.SampleRate * 2
			y := frequencyContent.Intensity / math.Pi * h

			imd.Push(pixel.V(x, 0))
			imd.Push(pixel.V(x, y))
			imd.Line(2)
			imd.Draw(win)
		}

		//imd.Draw(win)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
