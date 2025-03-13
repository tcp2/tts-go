// This example demonstrates how to list all available voices.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/difyz9/edge-tts-go/pkg/voices"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Get the list of voices
	voiceList, err := voices.ListVoices(ctx, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing voices: %v\n", err)
		os.Exit(1)
	}

	// Print the voices in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tGender\tLocale\tContentCategories\tVoicePersonalities")

	for _, voice := range voiceList {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			voice.ShortName,
			voice.Gender,
			voice.Locale,
			strings.Join(voice.VoiceTag.ContentCategories, ", "),
			strings.Join(voice.VoiceTag.VoicePersonalities, ", "),
		)
	}

	w.Flush()
}
