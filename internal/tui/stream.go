package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gi8lino/loggo/internal/ingest"
)

// frameMsg tells the model to process pending input and render a frame.
type frameMsg struct{}

// waitForBatch waits for the next ingested batch and drains ready batches.
func waitForBatch(stream <-chan ingest.Batch) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-stream
		if !ok {
			return ingest.Batch{EOF: true}
		}

		merged := msg

		for range 64 {
			select {
			case next, ok := <-stream:
				if !ok {
					merged.EOF = true
					return merged
				}

				if len(next.Lines) > 0 {
					merged.Lines = append(merged.Lines, next.Lines...)
				}
				if next.Err != nil {
					merged.Err = next.Err
				}
				if next.EOF {
					merged.EOF = true
					return merged
				}
			default:
				return merged
			}
		}

		return merged
	}
}

// tickFrame sends a frame message after the frame interval.
func tickFrame() tea.Cmd {
	return tea.Tick(frameInterval, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}
