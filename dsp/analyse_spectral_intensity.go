package dsp

import (
	"context"
	"github.com/strangedev/audiolight/audio"
)

type FrequencyBand struct {
	Name           string
	UpperFrequency float64
	LowerFrequency float64
}

type FrequencyBandIntensity struct {
	Band      FrequencyBand
	Intensity float64
}

type SpectralIntensityFrame = []FrequencyBandIntensity

type SpectralIntensityAnalysisOptions struct {
	Bands     []FrequencyBand
	FrameSize int
	FFTInterpreter
}

func NewSpectralIntensityAnalysisOptions(options audio.RecordingOptions) *SpectralIntensityAnalysisOptions {
	return &SpectralIntensityAnalysisOptions{
		FrameSize:      options.FrameSize,
		Bands:          []FrequencyBand{},
		FFTInterpreter: NewFFTInterpreter(options),
	}
}

func (options *SpectralIntensityAnalysisOptions) AddBand(name string, lowerFrequency float64, upperFrequency float64) *SpectralIntensityAnalysisOptions {
	options.Bands = append(
		options.Bands,
		FrequencyBand{
			Name:           name,
			LowerFrequency: lowerFrequency,
			UpperFrequency: upperFrequency,
		},
	)

	return options
}

func AnalyseSpectralIntensity(ctx context.Context, in <-chan FFTFrame, options *SpectralIntensityAnalysisOptions) <-chan SpectralIntensityFrame {
	out := make(chan SpectralIntensityFrame, audio.ChannelBufferSize)

	go func() {
		defer close(out)

		accumulatorFrame := make(SpectralIntensityFrame, len(options.Bands))
		bandSums := make([]float64, len(options.Bands))
		binsInBand := make([]float64, len(options.Bands))
		bandIndexForBin := make(map[int]int, options.FrameSize)

		for iBand, band := range options.Bands {
			accumulatorFrame[iBand] = FrequencyBandIntensity{
				Band:      options.Bands[iBand],
				Intensity: 0,
			}

			for iBin := 0; iBin < options.FrameSize-1; iBin++ {
				lowerBinFrequency := options.FFTInterpreter.GetFFTBinFrequency(iBin)
				upperBinFrequency := options.FFTInterpreter.GetFFTBinFrequency(iBin + 1)

				if upperBinFrequency >= band.UpperFrequency || lowerBinFrequency < band.LowerFrequency {
					continue
				}

				bandIndexForBin[iBin] = iBand
				binsInBand[iBand]++
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case frame := <-in:
				for iBin, sample := range frame {
					bandSums[bandIndexForBin[iBin]] += sample
				}
				for iBand, bandSum := range bandSums {
					accumulatorFrame[iBand].Intensity = bandSum / float64(options.FrameSize)
					bandSums[iBand] = 0

					outFrame := make(SpectralIntensityFrame, len(options.Bands))
					copy(outFrame, accumulatorFrame)

					out <- outFrame
				}

			}
		}
	}()

	return out
}
