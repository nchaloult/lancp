package app

import (
	"os"

	"github.com/nchaloult/lancp/pkg/net"
)

// Config stores input from command line arguments as well as configs set
// globally. It exposes lancp's core functionality, like running in send and
// receive mode.
type Config struct {
	// FilePath points to a file on disk that will be sent to the receiver.
	FilePath string

	// Port that lancp runs on locally, or listens for messages on locally.
	// Stored in the format ":0000".
	Port string

	// TLSPort is the port that lancp communicates via TLS on. Stored in the
	// format ":0000".
	TLSPort string
}

// NewSenderConfig returns a pointer to a new Config struct intended for use by
// lancp running in send mode.
func NewSenderConfig(filePath string, port, tlsPort int) (*Config, error) {
	// Make sure the file we want to send exists and we have access to it.
	if _, err := os.Stat(filePath); err != nil {
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

	return &Config{
		FilePath: filePath,
		Port:     portAsString,
		TLSPort:  tlsPortAsString,
	}, nil
}

// NewReceiverConfig returns a pointer to a new Config struct intended for use
// by lancp running in receive mode.
func NewReceiverConfig(port, tlsPort int) (*Config, error) {
	portAsString, err := net.GetPortAsString(port)
	if err != nil {
		return nil, err
	}
	tlsPortAsString, err := net.GetPortAsString(tlsPort)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:    portAsString,
		TLSPort: tlsPortAsString,
	}, nil
}
