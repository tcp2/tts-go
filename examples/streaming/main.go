// This example demonstrates how to use the edge-tts-go library with streaming and subtitles.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/difyz9/edge-tts-go/pkg/communicate"
	"github.com/difyz9/edge-tts-go/pkg/submaker"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Text to convert to speech
	text := "Hello, World! This is an example of how to use the edge-tts-go library with streaming and subtitles."

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

	// Create a SubMaker instance
	sm := submaker.NewSubMaker()

	// Open the output files
	audioFile, err := os.Create("output.mp3")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating audio file: %v\n", err)
		os.Exit(1)
	}
	defer audioFile.Close()

	subFile, err := os.Create("output.srt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating subtitle file: %v\n", err)
		os.Exit(1)
	}
	defer subFile.Close()

	// Stream the audio and metadata
	chunkChan, errChan := comm.Stream(ctx)

	// Process the chunks
	for chunk := range chunkChan {
		if chunk.Type == "audio" {
			_, err := audioFile.Write(chunk.Data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing audio data: %v\n", err)
				os.Exit(1)
			}
		} else if chunk.Type == "WordBoundary" {
			err := sm.Feed(chunk)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error feeding WordBoundary: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// Check for errors
	if err := <-errChan; err != nil {
		fmt.Fprintf(os.Stderr, "Error streaming: %v\n", err)
		os.Exit(1)
	}

	// Merge cues to reduce the number of cues
	err = sm.MergeCues(10) // 10 words per cue
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error merging cues: %v\n", err)
		os.Exit(1)
	}

	// Write the subtitles to the file
	_, err = fmt.Fprint(subFile, sm.GetSRT())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing subtitles: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Audio saved to output.mp3")
	fmt.Println("Subtitles saved to output.srt")
}
