package app

import (
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
	return nil
}
