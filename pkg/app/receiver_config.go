package app

import (
	"fmt"

	"github.com/nchaloult/lancp/pkg/cert"
	"github.com/nchaloult/lancp/pkg/handshake"
	"github.com/nchaloult/lancp/pkg/net"
)

// ReceiverConfig stores input from command line arguments, as well as configs
// that are set globally, for use when lancp is run with the "receive"
// subcommand.
type ReceiverConfig struct {
	port    string
	tlsPort string
}

// NewReceiverConfig returns a pointer to a new ReceiverConfig struct
// initialized with the provided arguments.
func NewReceiverConfig(port, tlsPort int) (*ReceiverConfig, error) {
	portAsString, err := net.GetPortAsString(port)
	if err != nil {
		return nil, err
	}
	tlsPortAsString, err := net.GetPortAsString(tlsPort)
	if err != nil {
		return nil, err
	}

	return &ReceiverConfig{
		port:    portAsString,
		tlsPort: tlsPortAsString,
	}, nil
}

// Run executes appropriate procedures when lancp is run with the "receive"
// subcommand. It completes an initial passphrase handshake with a sender,
// creates a self-signed TLS certificate for that sender to use, establishes a
// TLS connection with that sender, and receives a file.
func (c *ReceiverConfig) Run() error {
	conductor, err := handshake.NewReceiverConductor(
		c.port,
		handshakeTimeoutDuration,
	)
	if err != nil {
		return fmt.Errorf("failed to prepare for the lancp handshake: %v", err)
	}
	if err = conductor.ConductHandshake(); err != nil {
		return err
	}

	localAddr, err := net.GetPreferredOutboundAddr()
	if err != nil {
		return err
	}
	certificate, err := cert.GenerateSelfSignedCert(localAddr)
	if err != nil {
		return fmt.Errorf("failed to generate self-signed certificate: %v", err)
	}
	if err = cert.SendToSender(
		certificate,
		c.port,
		certTimeoutDuration,
	); err != nil {
		return fmt.Errorf("failed to send self-signed cert to sender: %v", err)
	}

	return nil
}
