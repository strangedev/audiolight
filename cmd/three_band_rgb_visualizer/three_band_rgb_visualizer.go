package main

import (
	"context"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/strangedev/audiolight/audio"
	"github.com/strangedev/audiolight/dsp"
	"image/color"
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
	downsampledFFTFrames := dsp.Downsample(ctx, fftFrames, recordingOptions.FrameSize, 3)

	spectralAnalysisOptions := dsp.NewSpectralIntensityAnalysisOptions(recordingOptions).
		AddBand("bass", 5, 1000).
		AddBand("middle", 1000, 10000).
		AddBand("treble", 10000, 20000)
	intensityFrames := dsp.AnalyseSpectralIntensity(ctx, downsampledFFTFrames, spectralAnalysisOptions)
	pressureMatchedFrames := dsp.DropFramesDynamically(ctx, intensityFrames)

	ticker := time.NewTicker(20 * time.Millisecond)

	scaleIntensityToUint8 := func(value float64) uint8 {
		return 20 + uint8(value*200*8)
	}

	for !win.Closed() {
		<-ticker.C

		frame := <-pressureMatchedFrames

		win.Clear(color.RGBA{
			scaleIntensityToUint8(frame[0].Intensity * 0.6),
			0,
			scaleIntensityToUint8(frame[2].Intensity * 2.3),
			1,
		})
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
