package main

import (
	"context"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/strangedev/audiolight/audio"
	"github.com/strangedev/audiolight/dsp"
	"golang.org/x/image/colornames"
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
	downsampledFFTFrames := dsp.Downsample(ctx, fftFrames, recordingOptions.FrameSize, 1)

	spectralAnalysisOptions := dsp.NewSpectralIntensityAnalysisOptions(recordingOptions).
		AddBand("bass", 5, 1000).
		AddBand("middle", 1000, 10000).
		AddBand("treble", 10000, 20000)
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

			x := w / float64(len(spectralAnalysisOptions.Bands)) * (float64(iBand) + 0.5)
			y := bandIntensity.Intensity * 6 * h

			imd.Push(pixel.V(x, 0))
			imd.Push(pixel.V(x, y))
			imd.Line(w / 3)
		}

		imd.Draw(win)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
