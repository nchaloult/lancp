package file

import (
	"fmt"
	"io"
	_net "net"
	"os"

	"github.com/alsm/ioprogress"
)

// Receiver is responsible for reading a payload from a network connection and
// writing it to a file on disk.
type Receiver struct {
	filePayloadBufSize int64
}

// NewReceiver returns a pointer to a new Receiver struct initialized with the
// provided fields.
func NewReceiver(filePayloadBufSize int64) (*Receiver, error) {
	if filePayloadBufSize <= 0 {
		return nil, fmt.Errorf("filePayloadBufSize should be greater than"+
			" zero, got: %d", filePayloadBufSize)
	}

	return &Receiver{filePayloadBufSize}, nil
}

// WritePayloadToFile reads a payload sent along the provided network connection
// and writes it to a file in bursts. It returns the number of bytes written to
// disk/received from a sender.
func (r *Receiver) WritePayloadToFile(
	file *os.File,
	fileSize int64,
	conn _net.Conn,
) (int64, error) {
	// progressReader is an io.Reader, and will write the progress of a read to
	// stdout in real time.
	progressReader := &ioprogress.Reader{
		Reader: conn,
		Size:   fileSize,
	}

	receivedBytes, err := io.Copy(file, progressReader)
	if err != nil {
		return 0, err
	}

	return receivedBytes, nil
}
