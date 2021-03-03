package io

import (
	"fmt"
	"io"
	"os"

	"github.com/alsm/ioprogress"
)

// getProgressReader returns a new Reader which, when read from, will display
// a progress bar to stderr.
func getProgressReader(size int64, reader io.Reader, barLen uint) io.Reader {
	// progressReader is an io.Reader, and will write the progress of a read to
	// stdout in real time.
	//
	// Inspired by the documented example on ioprogress.DrawTextFormatBar().
	bar := ioprogress.DrawTextFormatBar(int64(barLen))
	progressReader := &ioprogress.Reader{
		Reader: reader,
		Size:   size,
		// Draw to stderr so that this progress meter will always be displayed,
		// even if the user is piping or redirecting lancp's output someplace
		// else.
		DrawFunc: ioprogress.DrawTerminalf(
			os.Stderr,
			func(progress, total int64) string {
				return fmt.Sprintf(
					"%s %s",
					bar(progress, total),
					ioprogress.DrawTextFormatBytes(progress, total),
				)
			},
		),
	}

	return progressReader
}
