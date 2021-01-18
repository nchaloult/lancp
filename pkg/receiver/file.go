package receiver

import (
	"fmt"
	"io"
	_net "net"
	"os"

	"github.com/alsm/ioprogress"
)

// WritePayloadToFile reads a payload sent along the provided network connection
// and writes it to a file. It returns the number of bytes successfully received
// and written to disk.
func WritePayloadToFile(file *os.File, fileSize int64, conn _net.Conn) (int64, error) {
	// progressReader is an io.Reader, and will write the progress of a read to
	// stdout in real time.
	//
	// Inspired by the documented example on ioprogress.DrawTextFormatBar().
	bar := ioprogress.DrawTextFormatBar(40)
	progressReader := &ioprogress.Reader{
		Reader: conn,
		Size:   fileSize,
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

	return io.Copy(file, progressReader)
}
