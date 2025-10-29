package cli

import (
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
)

// ProgressBar wraps the progressbar library for consistent usage.
type ProgressBar struct {
	bar *progressbar.ProgressBar
}

// NewProgressBar creates a new progress bar with the given maximum value.
func NewProgressBar(max int, description string) *ProgressBar {
	bar := progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(100),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	return &ProgressBar{bar: bar}
}

// NewProgressBarQuiet creates a progress bar that writes to a custom writer.
func NewProgressBarQuiet(max int, description string, writer io.Writer) *ProgressBar {
	bar := progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(writer),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(100),
	)

	return &ProgressBar{bar: bar}
}

// Add increments the progress bar by the given amount.
func (pb *ProgressBar) Add(amount int) error {
	return pb.bar.Add(amount)
}

// Set sets the progress bar to a specific value.
func (pb *ProgressBar) Set(value int) error {
	return pb.bar.Set(value)
}

// Finish completes the progress bar.
func (pb *ProgressBar) Finish() error {
	return pb.bar.Finish()
}

// Clear clears the progress bar from the terminal.
func (pb *ProgressBar) Clear() error {
	return pb.bar.Clear()
}

// Describe updates the description of the progress bar.
func (pb *ProgressBar) Describe(description string) {
	pb.bar.Describe(description)
}
