package cert

import (
	_net "net"

	"github.com/nchaloult/lancp/pkg/net"
)

// ReceiveFromReceiver gets a TLS certificate from the receiver at the provided
// address through an insecure TCP connection.
//
// timeoutDuration is in seconds.
func ReceiveFromReceiver(addr _net.Addr, timeoutDuration uint) ([]byte, error) {
	conn, err := net.ConnectToTCPConn(addr, timeoutDuration)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return net.ReceiveMessage(conn, timeoutDuration)
}

// SendToSender establishes an insecure TCP connection with the sender and
// sends a TLS certificate.
//
// timeoutDuration is in seconds.
func SendToSender(
	certificate *SelfSignedCert,
	port string,
	timeoutDuration uint,
) error {
	ln, err := net.CreateTCPListener(port)
	if err != nil {
		return err
	}
	defer ln.Close()
	conn, err := net.EstablishConn(ln, timeoutDuration)
	if err != nil {
		return err
	}
	defer conn.Close()

	return net.SendMessage(certificate.Bytes, conn)
}
