package file

import (
	"fmt"
	"io"
	_net "net"
	"os"
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
) int64 {
	var receivedBytes int64
	for {
		if (fileSize - receivedBytes) < r.filePayloadBufSize {
			io.CopyN(file, conn, (fileSize - receivedBytes))
			conn.Read(make([]byte, (receivedBytes+int64(r.filePayloadBufSize))-fileSize))

			// Set receivedBytes so that the correct number of bytes received
			// will be displayed to the user.
			receivedBytes = fileSize

			break
		}

		io.CopyN(file, conn, int64(r.filePayloadBufSize))
		receivedBytes += int64(r.filePayloadBufSize)
	}

	return receivedBytes
}
