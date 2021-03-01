package net

import (
	"fmt"
	_net "net"
	"time"
)

// CreateTCPListener returns a TCP listener for this machine on the provided
// port. Caller is responsible for accepting incoming connection attempts on
// that listener.
//
// Port needs to look like a port string (i.e., ":xxxx" or ":xxxxx").
func CreateTCPListener(port string) (_net.Listener, error) {
	return _net.Listen("tcp", port)
}

// ConnectToTCPConn blocks as it attempts to connect to a TCP connection at the
// provided address. If it can't connect within the specified timeout duration,
// it returns an error that specifies such.
//
// timeoutDuration is in seconds.
//
// TODO: implement retry logic. Other machine may not have its listener ready to
// go the first time we reach out to connect.
func ConnectToTCPConn(addr _net.Addr, timeoutDuration uint) (_net.Conn, error) {
	connChan := make(chan _net.Conn, 1)
	errChan := make(chan error, 1)
	go func() {
		conn, err := _net.Dial("tcp", addr.String())
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
