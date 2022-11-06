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
	downsampledFFTFrames := dsp.Downsample(ctx, fftFrames, recordingOptions.FrameSize, 1)

	spectralAnalysisOptions := dsp.NewSpectralIntensityAnalysisOptions(recordingOptions).
		AddBand("bass", 10, 100).
		AddBand("middle", 100, 500).
		AddBand("treble", 500, 5000)
	intensityFrames := dsp.AnalyseSpectralIntensity(ctx, downsampledFFTFrames, spectralAnalysisOptions)
	pressureMatchedIntensityFrames := dsp.DropFramesDynamically(ctx, intensityFrames)

	ticker := time.NewTicker(20 * time.Millisecond)

	for !win.Closed() {
		<-ticker.C

		imd.Clear()
		win.Clear(colornames.Aliceblue)

		frame := <-pressureMatchedIntensityFrames

		for iBand, bandIntensity := range frame {
			imd.Color = colornames.Limegreen

			x := w / float64(len(spectralAnalysisOptions.Bands)) * float64(iBand)
			y := math.Log(1+bandIntensity.Intensity*10) * h

			imd.Push(pixel.V(x, y))
			imd.Circle(3, 6)
		}

		imd.Draw(win)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
