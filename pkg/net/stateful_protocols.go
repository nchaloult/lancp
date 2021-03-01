package net

import (
	"fmt"
	"io"
	_net "net"
	"time"
)

// EstablishConn blocks until it receives an attempt to establish a connection
// on the provided listener. If no attempt to begin the appropriate stateful
// protocol's handshake occurs within the specified timeout duration, it returns
// an error that specifies such.
//
// timeoutDuration is in seconds.
func EstablishConn(
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

// SendMessage sends the provided byte slice along the provided connection.
func SendMessage(message []byte, conn _net.Conn) error {
	_, err := conn.Write(message)
	return err
}

// ReceiveMessage blocks until it receives a message on the provided connection.
// If it receives a message, it returns that message and the address of the
// sender. If it doesn't receive a message within the specified timeout
// duration, it returns an error that specifies such.
//
// TODO: write logic for retries. Have this function accept a # of retries as a
// uint param.
func ReceiveMessage(
	conn _net.Conn,
	timeoutDuration uint,
) ([]byte, error) {
	msgChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
		msg, err := io.ReadAll(conn)
		if err != nil {
			errChan <- err
			return
		}

		msgChan <- msg
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

// FixedSizeMsg stores a byte slice and the number of bytes that were pushed to
// that slice. The caller/user can choose whether or not to discard the
// remaining bytes in the buffer. In most cases, they'll want to.
type FixedSizeMsg struct {
	Length int
	Bytes  []byte
}

// ReceiveMessageWithKnownSize blocks until it receives a message on the
// provided connection. It pushes that message into a fixed-size buffer of a
// specified length. If it doesn't receive a message within the specified
// timeout duration, it returns an error that specifies such.
//
// TODO: write logic for retries. Have this function accept a # of retries as a
// uint param.
func ReceiveMessageWithKnownSize(
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
