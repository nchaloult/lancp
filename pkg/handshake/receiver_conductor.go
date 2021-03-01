package handshake

import (
	"fmt"
	"log"
	"os"

	"github.com/nchaloult/lancp/pkg/input"
	"github.com/nchaloult/lancp/pkg/net"
	"github.com/nchaloult/lancp/pkg/passphrase"
)

// ReceiverConductor is responsible for executing the steps involved for a
// receiver in the lancp handshake process. It stores configurations for the
// handshake.
type ReceiverConductor struct {
	// capturer is the object responsible for prompting the user for
	// command-line input, and reading that input.
	capturer *input.Capturer

	// port is the UDP port that the handshake takes place on.
	port string

	// timeoutDuration is the number of seconds that the sender should wait for
	// responses from the sender before failing fast.
	timeoutDuration uint
}

// NewReceiverConductor returns a pointer to a new ReceiverConductor struct
// initialized with the provided parameters.
//
// timeoutDuration is in seconds.
//
// port needs to look like a port string (i.e., ":xxxx" or ":xxxxx").
func NewReceiverConductor(
	port string,
	timeoutDuration uint,
) (*ReceiverConductor, error) {
	capturer, err := input.NewCapturer("➜", "sender", os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}

	return &ReceiverConductor{capturer, port, timeoutDuration}, nil
}

// ConductHandshake executes the steps involved in the lancp handshake process.
// It listens for a UDP broadcast message from a potential sender, checks that
// sender's passphrase guess, reads in a passphrase guess from the user, and
// responds to the sender with that guess.
func (c *ReceiverConductor) ConductHandshake() error {
	// Display the expected passphrase for the receiver to send.
	expectedPassphrase := passphrase.Generate()
	log.Printf("Passphrase: %s\n", expectedPassphrase)

	// Receive broadcast message from sender, and check that the passphrase they
	// sent matches what we expect.
	conn, err := net.CreateUDPConn(c.port)
	if err != nil {
		return fmt.Errorf("failed to create a UDP connection for handshake: %v",
			err)
	}
	defer conn.Close()
	msg, err := net.ReceiveUDPMessage(conn, c.timeoutDuration, c.port)
	if err != nil {
		return fmt.Errorf("failed to receive broadcast message from sender: %v",
			err)
	}
	if msg.Payload != expectedPassphrase {
		return fmt.Errorf("got passphrase %q from sender, want %q",
			msg.Payload, expectedPassphrase)
	}

	// Ask the user to type in the passphrase that's displayed on the sender's
	// machine.
	input, err := c.capturer.CapturePassphrase()
	if err != nil {
		return fmt.Errorf("failed to capture passphrase input from user: %v",
			err)
	}

	// Send response with our passphrase guess to the sender.
	net.SendUDPMessage([]byte(input), conn, msg.ReturnAddr)

	return nil
}
