package app

import (
	"fmt"

	"github.com/nchaloult/lancp/pkg/cert"
	"github.com/nchaloult/lancp/pkg/file"
	"github.com/nchaloult/lancp/pkg/handshake"
	"github.com/nchaloult/lancp/pkg/io"
	"github.com/nchaloult/lancp/pkg/net"
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
	conductor, err := handshake.NewSenderConductor(
		c.port,
		handshakeTimeoutDuration,
	)
	if err != nil {
		return fmt.Errorf("failed to prepare for the lancp handshake: %v", err)
	}
	receiverAddr, err := conductor.ConductHandshake()
	if err != nil {
		return err
	}

	cert, err := cert.FetchFromReceiver(receiverAddr, certTimeoutDuration)
	if err != nil {
		return fmt.Errorf("failed to get TLS certificate from receiver: %v",
			err)
	}
	sender := file.NewSender(
		receiverAddr,
		c.tlsPort,
		c.filePath,
		cert,
		tlsTimeoutDuration,
		fileSendRetries,
	)
	err = sender.SendToReceiver()
	if err != nil {
		return fmt.Errorf("failed to send file to receiver: %v", err)
	}

	return nil
}
