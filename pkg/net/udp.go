package net

import (
	"fmt"
	_net "net"
	"time"
)

// TODO: make this user-configurable.
const minPassphrasePayloadBufSize = 32

// CreateUDPConn returns a UDP PacketConn for this machine on the provided port.
// Port needs to look like a port string (i.e., ":xxxx" or ":xxxxx").
func CreateUDPConn(port string) (_net.PacketConn, error) {
	return _net.ListenPacket("udp4", port)
}

// SendUDPMessage converts a provided message into a byte slice and sends it
// along the provided connection.
func SendUDPMessage(
	message string,
	conn _net.PacketConn,
	addr *_net.UDPAddr,
) error {
	_, err := conn.WriteTo([]byte(message), addr)
	return err
}

// UDPMessage stores a message's contents, as well as the address of the sender.
type UDPMessage struct {
	// Payload is a message's contents.
	Payload string

	// ReturnAddr is the address of the message's sender.
	ReturnAddr _net.Addr
}

// ReceiveUDPMessage blocks until it receives a UDP message on the provided
// connection. If it receives a message, it returns that message and the address
// of the sender. If it doesn't receive a message within the specified timeout
// duration, it returns a TimeoutError.
//
// It discards its own messages, like broadcast messages that get delivered to
// itself.
//
// timeoutDuration is in seconds.
//
// port needs to look like a port string (i.e., ":xxxx" or ":xxxxx").
func ReceiveUDPMessage(
	conn _net.PacketConn,
	timeoutDuration uint,
	port string,
) (*UDPMessage, error) {
	msgChan := make(chan *UDPMessage, 1)
	errChan := make(chan error, 1)
	go func() {
		// TODO: either pass the buffer size in as a param, or eventually make
		// this func a method on the HandshakeConductor struct (if you decide to
		// use something like it again).
		payloadBuf := make([]byte, minPassphrasePayloadBufSize)
		n, returnAddr, err := conn.ReadFrom(payloadBuf)
		if err != nil {
			errChan <- err
			return
		}

		// Discard our own broadcast message, and continue listening for one
		// more message.
		ourAddr, err := getLocalListeningAddr(port)
		if err != nil {
			errChan <- err
			return
		}
		if returnAddr.String() == ourAddr {
			n, returnAddr, err = conn.ReadFrom(payloadBuf)
		}

		payload := string(payloadBuf[:n])
		msgChan <- &UDPMessage{Payload: payload, ReturnAddr: returnAddr}
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
