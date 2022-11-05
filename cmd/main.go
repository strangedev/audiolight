package main

import (
	"context"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/strangedev/audiolight/audio"
	"golang.org/x/image/colornames"
	"os"
	"os/signal"
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

	recordingOptions := audio.RecordingOptions{SampleRate: 44100}
	frameChan, err := audio.RecordAudio(ctx, recordingOptions)
	if err != nil {
		panic(err)
	}

	fftChan, err := audio.FFT(ctx, frameChan)
	if err != nil {
		panic(err)
	}
	fftInterpreter := audio.NewFFTInterpreter(recordingOptions)

	for frame := range fftChan {
		if win.Closed() {
			break
		}

		win.Clear(colornames.Aliceblue)

		for i, frequencyContent := range fftInterpreter.GetFrequencyContent(frame) {
			imd.Color = colornames.Limegreen

			x := w / float64(len(frame)) * float64(i)
			y := frequencyContent.Intensity

			imd.Push(pixel.V(x, y))
			imd.Circle(2, 2)
		}

		imd.Draw(win)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
