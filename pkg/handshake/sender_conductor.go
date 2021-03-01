package handshake

import (
	"fmt"
	"log"
	_net "net"
	"os"

	"github.com/nchaloult/lancp/pkg/input"
	"github.com/nchaloult/lancp/pkg/net"
	"github.com/nchaloult/lancp/pkg/passphrase"
)

// SenderConductor is responsible for executing the steps involved for a sender
// in the lancp handshake process. It stores configurations for the handshake.
type SenderConductor struct {
	// capturer is the object responsible for prompting the user for
	// command-line input, and reading that input.
	capturer *input.Capturer

	// port is the UDP port that the handshake takes place on.
	port string

	// timeoutDuration is the number of seconds that the sender should wait for
	// responses from potential receivers before failing fast.
	timeoutDuration uint
}

// NewSenderConductor is responsible for executing the steps involved for a
// sender in the lancp handshake process. It stores configurations for the
// handshake.
//
// timeoutDuration is in seconds.
//
// port needs to look like a port string (i.e., ":xxxx" or ":xxxxx").
func NewSenderConductor(
	port string,
	timeoutDuration uint,
) (*SenderConductor, error) {
	capturer, err := input.NewCapturer("➜", "receiver", os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}

	return &SenderConductor{capturer, port, timeoutDuration}, nil
}

// ConductHandshake executes the steps involved in the lancp handshake process.
// It reads in a passphrase guess from the user, sends it in a UDP broadcast
// message, waits for a receiver to respond, and checks that receiver's
// passphrase guess.
//
// Returns the receiver's address so that we can attempt to establish a TCP
// connection with that address later.
func (c *SenderConductor) ConductHandshake() (_net.Addr, error) {
	// Ask the user to type in the passphrase that's displayed on the receiver's
	// machine.
	input, err := c.capturer.CapturePassphrase()
	if err != nil {
		return nil, fmt.Errorf("failed to capture passphrase input from user:"+
			" %v", err)
	}

	// Send UDP broadcast message to a receiver who's potentially listening.
	broadcastAddr, err := net.GetUDPBroadcastAddr(c.port)
	if err != nil {
		return nil, fmt.Errorf("failed to build UDP broadcast address: %v", err)
	}
	udpConn, err := net.CreateUDPConn(c.port)
	if err != nil {
		return nil, fmt.Errorf("failed to create a UDP connection for"+
			" handshake: %v", err)
	}
	defer udpConn.Close()
	net.SendUDPMessage(input, udpConn, broadcastAddr)

	// Display the expected passphrase for the receiver to send.
	expectedPassphrase := passphrase.Generate()
	log.Printf("Passphrase: %s\n", expectedPassphrase)

	// Receive response from receiver, and check that the passphrase they sent
	// matches what we expect.
	msg, err := net.ReceiveUDPMessage(udpConn, c.timeoutDuration, c.port)
	if err != nil {
		return nil, fmt.Errorf("failed to receive handshake response from"+
			" receiver: %v", err)
	}
	if msg.Payload != expectedPassphrase {
		return nil, fmt.Errorf("got passphrase %q from receiver, want %q",
			msg.Payload, expectedPassphrase)
	}

	return msg.ReturnAddr, nil
}
