package net

import (
	"crypto/tls"
	"fmt"
	_net "net"
	"time"
)

// CreateTLSListener returns a TLS listener for this machine on the provided
// port. Caller is responsible for accepting incoming connection attempts on
// that listener.
//
// Port needs to look like a port string (i.e., ":xxxx" or ":xxxxx").
func CreateTLSListener(cfg *tls.Config, port string) (_net.Listener, error) {
	return tls.Listen("tcp", port, cfg)
}

// ConnectToTLSConn blocks as it attempts to connect to a TLS connection at the
// provided address and with the provided TLS config. If it can't connect within
// the specified timeout duration, it returns an error that specifies such.
//
// timeoutDuration is in seconds.
//
// TODO: implement retry logic. Other machine may not have its listener ready to
// go the first time we reach out to connect.
func ConnectToTLSConn(
	addr string,
	config *tls.Config,
	timeoutDuration uint,
) (*tls.Conn, error) {
	connChan := make(chan *tls.Conn, 1)
	errChan := make(chan error, 1)
	go func() {
		conn, err := tls.Dial("tcp", addr, config)
		if err != nil {
			errChan <- err
			return
		}

		connChan <- conn
	}()

	select {
	case conn := <-connChan:
		return conn, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(time.Duration(timeoutDuration) * time.Second):
		return nil, fmt.Errorf("timed out after %d seconds", timeoutDuration)
	}
}
