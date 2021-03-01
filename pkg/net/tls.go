package net

import (
	"crypto/tls"
	"fmt"
	"time"
)

// ConnectToTLSConn blocks as it attempts to connect to a TLS connection at the
// provided address and with the provided TLS config. If it can't connect within
// the specified timeout duration, it returns an error that specifies such.
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

// SendTLSMessage writes a byte string to the provided connection.
func SendTLSMessage(message []byte, conn *tls.Conn) error {
	_, err := conn.Write(message)
	return err
}
