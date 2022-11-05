package main

import (
	"context"
	"fmt"
	"github.com/strangedev/audiolight/audio"
	"os"
	"os/signal"
)

func main() {
	fmt.Println("Recording.  Press Ctrl-C to stop.")

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	frameChan, err := audio.RecordAudio(ctx, audio.RecordingOptions{SampleRate: 16000})
	if err != nil {
		panic(err)
	}

	for frame := range frameChan {
		fmt.Printf("Frame: %+v\n", frame)
	}
}
