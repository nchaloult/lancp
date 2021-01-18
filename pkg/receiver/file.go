package receiver

import (
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
	progressReader := &ioprogress.Reader{
		Reader: conn,
		Size:   fileSize,
	}

	return io.Copy(file, progressReader)
}
