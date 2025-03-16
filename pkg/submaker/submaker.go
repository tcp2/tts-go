// Package submaker is used to generate subtitles from WordBoundary and SentenceBoundary events.
package submaker

import (
	"fmt"
	"strings"
	"time"

	"github.com/difyz9/edge-tts-go/pkg/types"
)

// SubMaker is used to generate subtitles from WordBoundary and SentenceBoundary messages.
type SubMaker struct {
	cues []Subtitle
}

// Subtitle represents a subtitle cue.
type Subtitle struct {
	Index   int
	Start   time.Duration
	End     time.Duration
	Content string
}

// NewSubMaker creates a new SubMaker.
func NewSubMaker() *SubMaker {
	return &SubMaker{
		cues: []Subtitle{},
	}
}

// Feed feeds a WordBoundary or SentenceBoundary message to the SubMaker.
func (sm *SubMaker) Feed(msg types.TTSChunk) error {
	if msg.Type != "WordBoundary" && msg.Type != "SentenceBoundary" {
		return fmt.Errorf("invalid message type, expected 'WordBoundary' or 'SentenceBoundary', got '%s'", msg.Type)
	}

	sm.cues = append(sm.cues, Subtitle{
		Index:   len(sm.cues) + 1,
		Start:   time.Duration(msg.Offset / 10) * time.Microsecond,
		End:     time.Duration((msg.Offset + msg.Duration) / 10) * time.Microsecond,
		Content: msg.Text,
	})

	return nil
}

// MergeCues merges cues to reduce the number of cues.
func (sm *SubMaker) MergeCues(words int) error {
	if words <= 0 {
		return fmt.Errorf("invalid number of words to merge, expected > 0, got %d", words)
	}

	if len(sm.cues) == 0 {
		return nil
	}

	newCues := []Subtitle{}
	currentCue := sm.cues[0]

	for _, cue := range sm.cues[1:] {
		if len(strings.Fields(currentCue.Content)) < words {
			currentCue.End = cue.End
			currentCue.Content = currentCue.Content + " " + cue.Content
		} else {
			newCues = append(newCues, currentCue)
			currentCue = cue
		}
	}

	newCues = append(newCues, currentCue)

	// Update indices
	for i := range newCues {
		newCues[i].Index = i + 1
	}

	sm.cues = newCues
	return nil
}

// GetSRT returns the SRT formatted subtitles from the SubMaker.
func (sm *SubMaker) GetSRT() string {
	var sb strings.Builder

	for _, cue := range sm.cues {
		// Format: "00:00:00,000 --> 00:00:00,000"
		startStr := formatDuration(cue.Start)
		endStr := formatDuration(cue.End)

		sb.WriteString(fmt.Sprintf("%d\n", cue.Index))
		sb.WriteString(fmt.Sprintf("%s --> %s\n", startStr, endStr))
		sb.WriteString(fmt.Sprintf("%s\n\n", cue.Content))
	}

	return sb.String()
}

// formatDuration formats a duration as "00:00:00,000".
func formatDuration(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	d -= s * time.Second
	ms := d / time.Millisecond

	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

// String returns the SRT formatted subtitles from the SubMaker.
func (sm *SubMaker) String() string {
	return sm.GetSRT()
}
