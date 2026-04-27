package utils

import (
	"github.com/schollz/progressbar/v3"
)

// NewProgressBar creates a new progress bar for downloading
func NewProgressBar(maxBytes int64, description string) *progressbar.ProgressBar {
	return progressbar.DefaultBytes(
		maxBytes,
		description,
	)
}
