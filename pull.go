package main

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/rs/zerolog"
)

// https://riptutorial.com/docker/example/31980/image-pulling-with-progress-bars--written-in-go
type pullEvent struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
	Progress       string `json:"progress,omitempty"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

func handleImagePull(resp io.ReadCloser, logger zerolog.Logger) error {
	decoder := json.NewDecoder(resp)
	var event *pullEvent

	for {
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		// Check if the line is one of the final two ones
		if strings.HasPrefix(event.Status, "Digest:") || strings.HasPrefix(event.Status, "Status:") {
			logger.Info().
				Str("status", event.Status).
				Msg("image pull update")
			continue
		}

		if event.Status == "Pull complete" {
			logger.Info().Str("status", event.Status).Msg("image pull completed")
		} else {
			logger.Info().Str("status", event.Status).Str("progress", event.Progress).Msg("image pull")
		}

	}

	return nil
}
