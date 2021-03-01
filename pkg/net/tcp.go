package net

import (
	"fmt"
	"io"
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

// EstablishTCPConn blocks until it receives an attempt to establish a TCP
// connection on the provided listener. If no attempt to begin the TCP handshake
// occurs within the specified timeout duration, it returns an error that
// specifies such.
//
// timeoutDuration is in seconds.
func EstablishTCPConn(
	ln _net.Listener,
	timeoutDuration uint,
) (_net.Conn, error) {
	connChan := make(chan _net.Conn, 1)
	errChan := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
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

// SendTCPMessage sends the provided byte slice along the provided connection.
func SendTCPMessage(message []byte, conn _net.Conn) error {
	_, err := conn.Write(message)
	return err
}

// ReceiveTCPMessage blocks until it receives a TCP message on the provided
// connection. If it receives a message, it returns that message and the address
// of the sender. If it doesn't receive a message within the specified timeout
// duration, it returns an error that specifies such.
//
// TODO: write logic for retries. Have this function accept a # of retries as a
// uint param.
func ReceiveTCPMessage(
	conn _net.Conn,
	timeoutDuration uint,
) ([]byte, error) {
	certChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
		cert, err := io.ReadAll(conn)
		if err != nil {
			errChan <- err
			return
		}

		certChan <- cert
	}()

	select {
	case cert := <-certChan:
		return cert, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(time.Duration(timeoutDuration) * time.Second):
		return nil, fmt.Errorf("timed out after %d seconds", timeoutDuration)
	}
}
