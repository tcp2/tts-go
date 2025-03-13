// This is a simple example of how to use the edge-tts-go library.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/difyz9/edge-tts-go/pkg/communicate"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Text to convert to speech
	text := "Hello, World! This is a simple example of how to use the edge-tts-go library."

	// Voice to use
	voice := "en-US-GuyNeural"

	// Create a new Communicate instance
	comm, err := communicate.NewCommunicate(
		text,
		voice,
		"+0%",  // rate
		"+0%",  // volume
		"+0Hz", // pitch
		"",     // proxy
		10,     // connectTimeout
		60,     // receiveTimeout
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Communicate instance: %v\n", err)
		os.Exit(1)
	}

	// Save the audio to a file
	err = comm.Save(ctx, "output.mp3", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving audio: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Audio saved to output.mp3")
}
