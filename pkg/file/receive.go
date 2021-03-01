package file

import (
	"fmt"
	"log"

	"github.com/nchaloult/lancp/pkg/cert"
	"github.com/nchaloult/lancp/pkg/net"
)

// Sender is responsible for sending a file to a receiver over a TLS connection.
type Sender struct {
	addr            string
	certificate     []byte
	filePath        string
	timeoutDuration uint
	numRetries      uint
}

// NewSender returns a new Sender object initialized with the provided
// parameters.
//
// port is the TLS port. It needs to look like a port string (i.e., ":xxxx" or
// ":xxxxx").
func NewSender(
	addr, port, filePath string,
	certificate []byte,
	timeoutDuration, numRetries uint,
) *Sender {
	return &Sender{
		// Build TLS address.
		addr: net.GetTLSAddress(addr, port),

		certificate:     certificate,
		filePath:        filePath,
		timeoutDuration: timeoutDuration,
		numRetries:      numRetries,
	}
}

// SendToReceiver sends a file to the receiver at the provided address along a
// TLS connection. It builds a TLS config struct with necessary information to
// establish a TLS connection, establishes that connection, sends the name and
// size of the file at the provided path, and sends the file's contents.
func (s *Sender) SendToReceiver() error {
	// Connect to the receiver's TLS conn with the provided cert.
	tlsCfg := cert.GetSenderTLSConfig(s.certificate)
	conn, err := net.ConnectToTLSConn(s.addr, tlsCfg, s.timeoutDuration)
	if err != nil {
		return fmt.Errorf("failed to establish TLS connection with receiver:"+
			" %v", err)
	}
	defer conn.Close()

	// TODO: stubbed.
	log.Println(conn)
	return nil
}
