package net

import (
	"fmt"
	"log"
	_net "net"
	"os"

	"github.com/nchaloult/lancp/pkg/input"
)

const (
	minPassphrasePayloadBufSize = 32
	maxPassphrasePayloadBufSize = 1024
)

// HandshakeConductor carries out the choreographed procedures for lancp's
// device discovery and identity validation handshake.
//
// TODO: write explanatory comments for each of the fields.
type HandshakeConductor struct {
	udpConn                  _net.PacketConn
	passphrasePayloadBufSize uint
	expectedPassphrase       string
	ourAddr                  string
}

// NewHandshakeConductor returns a pointer to a new HandshakeConductor
// initialized with the provided fields.
func NewHandshakeConductor(
	udpConn _net.PacketConn,
	passphrasePayloadBufSize uint,
	expectedPassphrase string,
	ourAddr string,
) (*HandshakeConductor, error) {
	if passphrasePayloadBufSize < minPassphrasePayloadBufSize ||
		passphrasePayloadBufSize > maxPassphrasePayloadBufSize {
		return nil, fmt.Errorf("passphrasePayloadBufSize should be in the"+
			" range [%d, %d], got: %d", minPassphrasePayloadBufSize,
			maxPassphrasePayloadBufSize, passphrasePayloadBufSize)
	}

	return &HandshakeConductor{
		udpConn,
		passphrasePayloadBufSize,
		expectedPassphrase,
		ourAddr,
	}, nil
}

// PerformHandshakeAsReceiver waits for a sender to reach out and attempt to
// begin a handshake, compares the passphrase it sent with the expected
// passphrase, and sends a response with what it believes is the sender's
// passphrase.
//
// All of the messages exchanged between a sender and receiver during this
// handshake are sent over UDP, which means that the caller may close the
// net.PacketConn it created the HandshakeConductor with as soon as this method
// returns.
func (hc *HandshakeConductor) PerformHandshakeAsReceiver() error {
	// Listen for a broadcast message from a sender.
	senderPayload, senderAddr, err := hc.getPassphraseFromMessage()
	if err != nil {
		return fmt.Errorf("failed to read broadcast message from sender: %v",
			err)
	}

	// Check the sender's passphrase.
	if senderPayload != hc.expectedPassphrase {
		return fmt.Errorf("got passphrase: %q from sender, want %q",
			senderPayload, hc.expectedPassphrase)
	}

	// Capture user input for the passphrase the sender is presenting.
	capturer, err := input.NewCapturer("➜", false, os.Stdin, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to create a new Capturer: %v", err)
	}
	userInput, err := capturer.CapturePassphrase()
	if err != nil {
		return err
	}

	// Send response message to sender with its guess for the sender's
	// passphrase.
	_, err = hc.udpConn.WriteTo([]byte(userInput), senderAddr)
	if err != nil {
		return fmt.Errorf("failed to send response message to sender: %v", err)
	}

	return nil
}

// PerformHandshakeAsSender begins the lancp handshake process. It reads in a
// passphrase guess from the user, sends it in a UDP broadcast message, waits
// for a response from a receiver, and checks that receiver's passphrase guess.
//
// All of the messages exchanged between a sender and receiver during this
// handshake are sent over UDP, which means that the caller may close the
// net.PacketConn it created the HandshakeConductor with as soon as this method
// returns.
//
// Returns the receiver's address so that we can attempt to establish a TCP
// connection with that address later.
//
// TODO: should broadcastUDPAddr be passed in as a parameter like this? Is
// HandshakeConductor responsible for too much? Should we have one version for
// the receiver to use, and another for the sender?
func (hc *HandshakeConductor) PerformHandshakeAsSender(
	broadcastUDPAddr *_net.UDPAddr,
) (_net.Addr, error) {
	// Capture user input for the passphrase the receiver is presenting.
	capturer, err := input.NewCapturer("➜", true, os.Stdin, os.Stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new Capturer: %v", err)
	}
	userInput, err := capturer.CapturePassphrase()
	if err != nil {
		return nil, fmt.Errorf("failed to capture passphrase input from user:"+
			" %v", err)
	}

	_, err = hc.udpConn.WriteTo([]byte(userInput), broadcastUDPAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to send UDP broadcast message: %v", err)
	}

	// Display the generated passphrase for the receiver to send.
	log.Printf("Passphrase: %s\n", hc.expectedPassphrase)

	// Listen for a broadcast message from a receiver.
	receiverPayload, returnAddr, err := hc.getPassphraseFromMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read broadcast message from sender: %v",
			err)
	}

	// Check the receiver's passphrase.
	if receiverPayload != hc.expectedPassphrase {
		return nil, fmt.Errorf("got passphrase: %q from receiver, want %q",
			receiverPayload, hc.expectedPassphrase)
	}

	return returnAddr, nil
}

// getPassphraseFromMessage blocks until it receives a UDP message with a
// passphrase guess as its payload, and returns that payload. It discards
// its own messages, like broadcast messages that get delivered to itself.
func (hc *HandshakeConductor) getPassphraseFromMessage() (string, _net.Addr, error) {
	payloadBuf := make([]byte, hc.passphrasePayloadBufSize)
	n, returnAddr, err := hc.udpConn.ReadFrom(payloadBuf)
	if err != nil {
		return "", nil, err
	}

	if returnAddr.String() == hc.ourAddr {
		// Discard our own broadcast message and continue listening for one more
		// message.
		n, returnAddr, err = hc.udpConn.ReadFrom(payloadBuf)
	}

	payload := string(payloadBuf[:n])
	return payload, returnAddr, nil
}
