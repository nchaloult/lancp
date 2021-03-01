package file

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/nchaloult/lancp/pkg/cert"
	"github.com/nchaloult/lancp/pkg/io"
	"github.com/nchaloult/lancp/pkg/net"
)

const (
	// TODO: Is this an okay size for this buffer? How big could it ever get?
	// https://www.ibm.com/support/knowledgecenter/SSEQVQ_8.1.10/client/c_cmd_filespecsyntax.html
	nameBufLen = 1024

	// Combo of answers from https://stackoverflow.com/questions/35371385/how-can-i-convert-an-int64-into-a-byte-array-in-go
	sizeBufLen = binary.MaxVarintLen64
)

// ReceiveFromSender receives a file from the sender along a TLS connection and
// saves it to disk. It builds a TLS config struct with necessary information to
// establish a TLS connection, establishes that connection, receives the file's
// name and size, then the file's contents, and saves it to disk.
func ReceiveFromSender(
	certificate *cert.SelfSignedCert,
	port string,
	timeoutDuration uint,
) error {
	// Stand up a TLS conn.
	cfg, err := cert.GetReceiverTLSConfig(certificate)
	if err != nil {
		return fmt.Errorf("failed to prepare for TLS: %v", err)
	}
	ln, err := net.CreateTLSListener(cfg, port)
	if err != nil {
		return fmt.Errorf("failed to create TLS listener: %v", err)
	}
	defer ln.Close()
	conn, err := net.EstablishTLSConn(ln, timeoutDuration)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Receive the file's name and size from the sender.
	nameBuf, err := net.ReceiveTLSMessageWithKnownSize(
		nameBufLen,
		conn,
		timeoutDuration,
	)
	if err != nil {
		return fmt.Errorf("failed to receive file name from sender: %v", err)
	}
	name := string(nameBuf.Bytes[:nameBuf.Length])
	sizeBuf, err := net.ReceiveTLSMessageWithKnownSize(
		sizeBufLen,
		conn,
		timeoutDuration,
	)
	if err != nil {
		return fmt.Errorf("failed to receive file size from sender: %v", err)
	}
	size, bytesRead := binary.Varint(sizeBuf.Bytes)
	if bytesRead == 0 {
		if size == 0 {
			return errors.New("failed to receive file size from sender:" +
				" buffer too small on our end (this should never happen, but" +
				" it happened lol)")
		}
		return errors.New("failed to receive file size from sender: size is" +
			" larger than the max value of a 64-bit integer (this should" +
			" never happen, but it happened lol)")
	}

	file, err := io.CreateNewFileOnDisk(name)
	if err != nil {
		return fmt.Errorf("failed to create a new file on disk: %v", err)
	}
	defer file.Close()

	return io.ReceiveFileFromConn(file, size, conn)
}
