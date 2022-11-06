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

const w, h = float64(1024), float64(512)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
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
	binCount := fftInterpreter.GetBinCount()

	ticker := time.NewTicker(20 * time.Millisecond)

	for !win.Closed() {
		<-ticker.C

		imd.Clear()
		win.Clear(colornames.Aliceblue)

		frame := <-pressureMatchedFFTFrames

		for i, frequencyContent := range fftInterpreter.GetFrequencyContent(frame) {
			imd.Color = colornames.Limegreen

			x := w / float64(binCount) * float64(i)
			y := math.Log(1+frequencyContent.Intensity*10) * h

			imd.Push(pixel.V(x, y))
			imd.Circle(1, 2)
		}

		imd.Draw(win)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
