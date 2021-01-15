package net

import (
	"fmt"
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
}

// NewHandshakeConductor returns a pointer to a new HandshakeConductor
// initialized with the provided fields.
func NewHandshakeConductor(
	udpConn _net.PacketConn,
	passphrasePayloadBufSize uint,
	expectedPassphrase string,
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
	senderPayload, senderAddr, err := hc.getPassphraseFromSender()
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

// getPassphraseFromSender performs the initial step of the handshake as the
// receiver. It blocks until a sender sends a UDP broadcast message with a
// passphrase as its payload, and returns that payload.
func (hc *HandshakeConductor) getPassphraseFromSender() (string, _net.Addr, error) {
	payloadBuf := make([]byte, hc.passphrasePayloadBufSize)
	n, senderAddr, err := hc.udpConn.ReadFrom(payloadBuf)
	if err != nil {
		return "", nil, err
	}

	payload := string(payloadBuf[:n])
	return payload, senderAddr, nil
}
