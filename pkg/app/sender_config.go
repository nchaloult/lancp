package app

import (
	"fmt"
	"log"
	"os"

	"github.com/nchaloult/lancp/pkg/input"
	"github.com/nchaloult/lancp/pkg/io"
	"github.com/nchaloult/lancp/pkg/net"
	"github.com/nchaloult/lancp/pkg/passphrase"
)

// TODO: read these in from some config file, env vars, or something.
const (
	timeoutDuration = 60
)

// SenderConfig stores input from command line arguments, as well as configs
// that are set globally, for use when lancp is run with the "send" subcommand.
type SenderConfig struct {
	filePath string
	port     string
	tlsPort  string
}

// NewSenderConfig returns a pointer to a new SenderConfig struct initialized
// with the provided arguments.
func NewSenderConfig(filePath string, port, tlsPort int) (*SenderConfig, error) {
	if err := io.IsFileAccessible(filePath); err != nil {
		return nil, err
	}

	portAsString, err := net.GetPortAsString(port)
	if err != nil {
		return nil, err
	}
	tlsPortAsString, err := net.GetPortAsString(tlsPort)
	if err != nil {
		return nil, err
	}

	return &SenderConfig{
		filePath: filePath,
		port:     portAsString,
		tlsPort:  tlsPortAsString,
	}, nil
}

// Run executes appropriate procedures when lancp is run with the "send"
// subcommand. It completes an initial passphrase handshake with a receiver,
// receives a TLS certificate from that receiver, establishes a TLS connection
// with that certificate, and sends a file.
func (c *SenderConfig) Run() error {
	// Ask the user to type in the passphrase that's displayed on the receiver's
	// machine.
	capturer, err := input.NewCapturer("➜", "receiver", os.Stdin, os.Stdout)
	if err != nil {
		return err
	}
	input, err := capturer.CapturePassphrase()
	if err != nil {
		return fmt.Errorf("failed to capture passphrase input from user: %v",
			err)
	}

	// Send UDP broadcast message to a receiver who's potentially listening.
	broadcastAddr, err := net.GetUDPBroadcastAddr(c.port)
	if err != nil {
		return fmt.Errorf("failed to build UDP broadcast address: %v", err)
	}
	udpConn, err := net.CreateUDPConn(c.port)
	if err != nil {
		return fmt.Errorf("failed to create a UDP connection for handshake: %v",
			err)
	}
	net.SendUDPMessage(input, udpConn, broadcastAddr)

	// Display the expected passphrase for the receiver to send.
	expectedPassphrase := passphrase.Generate()
	log.Printf("Passphrase: %s\n", expectedPassphrase)

	// Receive response from receiver, and check that the passphrase they sent
	// matches what we expect.
	msg, err := net.ReceiveUDPMessage(
		udpConn,
		timeoutDuration,
		c.port,
	)
	if err != nil {
		return fmt.Errorf("failed to receive handshake response from receiver:"+
			" %v", err)
	}
	if msg.Payload != expectedPassphrase {
		return fmt.Errorf("got passphrase %q from receiver, want %q",
			msg.Payload, expectedPassphrase)
	}

	// TODO: stubbed.
	log.Println(msg.ReturnAddr)
	return nil
}
