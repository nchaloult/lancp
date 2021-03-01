package cert

import (
	_net "net"

	"github.com/nchaloult/lancp/pkg/net"
)

// FetchFromReceiver gets a TLS certificate from the receiver at the provided
// address through an insecure TCP connection.
//
// timeoutDuration is in seconds.
func FetchFromReceiver(addr _net.Addr, timeoutDuration uint) ([]byte, error) {
	conn, err := net.ConnectToTCPConn(addr, timeoutDuration)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return net.ReceiveTCPMessage(conn, timeoutDuration)
}
