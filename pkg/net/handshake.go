package net

import (
	"fmt"
	_net "net"
	"time"
)

const (
	minPassphrasePayloadBufSize = 32
	maxPassphrasePayloadBufSize = 1024
)

// // HandshakeConductor carries out the choreographed procedures for lancp's
// // device discovery and identity validation handshake.
// //
// // TODO: write explanatory comments for each of the fields.
// type HandshakeConductor struct {
// 	udpConn                  _net.PacketConn
// 	passphrasePayloadBufSize uint
// 	expectedPassphrase       string
// 	ourAddr                  string
// }

// // NewHandshakeConductor returns a pointer to a new HandshakeConductor
// // initialized with the provided fields.
// func NewHandshakeConductor(
// 	udpConn _net.PacketConn,
// 	passphrasePayloadBufSize uint,
// 	expectedPassphrase string,
// 	ourAddr string,
// ) (*HandshakeConductor, error) {
// 	if passphrasePayloadBufSize < minPassphrasePayloadBufSize ||
// 		passphrasePayloadBufSize > maxPassphrasePayloadBufSize {
// 		return nil, fmt.Errorf("passphrasePayloadBufSize should be in the"+
// 			" range [%d, %d], got: %d", minPassphrasePayloadBufSize,
// 			maxPassphrasePayloadBufSize, passphrasePayloadBufSize)
// 	}

// 	return &HandshakeConductor{
// 		udpConn,
// 		passphrasePayloadBufSize,
// 		expectedPassphrase,
// 		ourAddr,
// 	}, nil
// }

// // PerformHandshakeAsReceiver waits for a sender to reach out and attempt to
// // begin a handshake, compares the passphrase it sent with the expected
// // passphrase, and sends a response with what it believes is the sender's
// // passphrase.
// //
// // All of the messages exchanged between a sender and receiver during this
// // handshake are sent over UDP, which means that the caller may close the
// // net.PacketConn it created the HandshakeConductor with as soon as this method
// // returns.
// func (hc *HandshakeConductor) PerformHandshakeAsReceiver() error {
// 	// Display the expected passphrase for the sender to send.
// 	log.Printf("Passphrase: %s\n", hc.expectedPassphrase)

// 	// Listen for a broadcast message from a sender.
// 	senderPayload, senderAddr, err := hc.getPassphraseFromMessage()
// 	if err != nil {
// 		return fmt.Errorf("failed to read broadcast message from sender: %v",
// 			err)
// 	}

// 	// Check the sender's passphrase.
// 	if senderPayload != hc.expectedPassphrase {
// 		return fmt.Errorf("got passphrase: %q from sender, want %q",
// 			senderPayload, hc.expectedPassphrase)
// 	}

// 	// Capture user input for the passphrase the sender is presenting.
// 	capturer, err := input.NewCapturer("➜", false, os.Stdin, os.Stdout)
// 	if err != nil {
// 		return fmt.Errorf("failed to create a new Capturer: %v", err)
// 	}
// 	userInput, err := capturer.CapturePassphrase()
// 	if err != nil {
// 		return err
// 	}

// 	// Send response message to sender with its guess for the sender's
// 	// passphrase.
// 	_, err = hc.udpConn.WriteTo([]byte(userInput), senderAddr)
// 	if err != nil {
// 		return fmt.Errorf("failed to send response message to sender: %v", err)
// 	}

// 	return nil
// }

// // PerformHandshakeAsSender begins the lancp handshake process. It reads in a
// // passphrase guess from the user, sends it in a UDP broadcast message, waits
// // for a response from a receiver, and checks that receiver's passphrase guess.
// //
// // All of the messages exchanged between a sender and receiver during this
// // handshake are sent over UDP, which means that the caller may close the
// // net.PacketConn it created the HandshakeConductor with as soon as this method
// // returns.
// //
// // Returns the receiver's address so that we can attempt to establish a TCP
// // connection with that address later.
// //
// // TODO: should broadcastUDPAddr be passed in as a parameter like this? Is
// // HandshakeConductor responsible for too much? Should we have one version for
// // the receiver to use, and another for the sender?
// func (hc *HandshakeConductor) PerformHandshakeAsSender(
// 	broadcastUDPAddr *_net.UDPAddr,
// ) (_net.Addr, error) {
// 	// Capture user input for the passphrase the receiver is presenting.
// 	capturer, err := input.NewCapturer("➜", true, os.Stdin, os.Stdout)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create a new Capturer: %v", err)
// 	}
// 	userInput, err := capturer.CapturePassphrase()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to capture passphrase input from user:"+
// 			" %v", err)
// 	}

// 	_, err = hc.udpConn.WriteTo([]byte(userInput), broadcastUDPAddr)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to send UDP broadcast message: %v", err)
// 	}

// 	// Display the expected passphrase for the receiver to send.
// 	log.Printf("Passphrase: %s\n", hc.expectedPassphrase)

// 	// Listen for a broadcast message from a receiver.
// 	receiverPayload, returnAddr, err := hc.getPassphraseFromMessage()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read broadcast message from sender: %v",
// 			err)
// 	}

// 	// Check the receiver's passphrase.
// 	if receiverPayload != hc.expectedPassphrase {
// 		return nil, fmt.Errorf("got passphrase: %q from receiver, want %q",
// 			receiverPayload, hc.expectedPassphrase)
// 	}

// 	return returnAddr, nil
// }

// // getPassphraseFromMessage blocks until it receives a UDP message with a
// // passphrase guess as its payload, and returns that payload. It discards
// // its own messages, like broadcast messages that get delivered to itself.
// func (hc *HandshakeConductor) getPassphraseFromMessage() (string, _net.Addr, error) {
// 	payloadBuf := make([]byte, hc.passphrasePayloadBufSize)
// 	n, returnAddr, err := hc.udpConn.ReadFrom(payloadBuf)
// 	if err != nil {
// 		return "", nil, err
// 	}

// 	if returnAddr.String() == hc.ourAddr {
// 		// Discard our own broadcast message and continue listening for one more
// 		// message.
// 		n, returnAddr, err = hc.udpConn.ReadFrom(payloadBuf)
// 	}

// 	payload := string(payloadBuf[:n])
// 	return payload, returnAddr, nil
// }

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
