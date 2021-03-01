package net

import (
	"crypto/tls"
	"fmt"
	"io"
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

// EstablishTLSConn blocks until it receives an attempt to establish a TLS
// connection on the provided listener. If no attempt to begin the TLS handshake
// occurs within the specified timeout duration, it returns an error that
// specifies such.
//
// timeoutDuration is in seconds.
func EstablishTLSConn(
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

// SendTLSMessage writes a byte string to the provided connection.
func SendTLSMessage(message []byte, conn _net.Conn) error {
	_, err := conn.Write(message)
	return err
}

// ReceiveTLSMessage blocks until it receives a TLS message on the provided
// connection. If it receives a message, it returns that message and the address
// of the sender. If it doesn't receive a message within the specified timeout
// duration, it returns an error that specifies such.
//
// TODO: write logic for retries. Have this function accept a # of retries as a
// uint param.
func ReceiveTLSMessage(
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

// FixedSizeMsg stores a byte slice and the number of bytes that were pushed to
// that slice. The caller/user can choose whether or not to discard the
// remaining bytes in the buffer. In most cases, they'll want to.
type FixedSizeMsg struct {
	Length int
	Bytes  []byte
}

// ReceiveTLSMessageWithKnownSize blocks until it receives a TLS message on the
// provided connection. It reads that message into a fixed-size buffer of a
// specified length. It returns the number of bytes pushed into that buffer.
//
// If it receives a message, it returns that message and the address of the
// sender. If it doesn't receive a message within the specified timeout
// duration, it returns an error that specifies such.
//
// TODO: write logic for retries. Have this function accept a # of retries as a
// uint param.
func ReceiveTLSMessageWithKnownSize(
	size uint,
	conn _net.Conn,
	timeoutDuration uint,
) (*FixedSizeMsg, error) {
	msgChan := make(chan *FixedSizeMsg, 1)
	errChan := make(chan error, 1)
	go func() {
		msgBuf := make([]byte, size)
		n, err := conn.Read(msgBuf)
		if err != nil {
			errChan <- err
			return
		}

		msgChan <- &FixedSizeMsg{
			Length: n,
			Bytes:  msgBuf,
		}
	}()

	select {
	case msg := <-msgChan:
		return msg, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(time.Duration(timeoutDuration) * time.Second):
		return nil, fmt.Errorf("timed out after %d seconds", timeoutDuration)
	}
}
