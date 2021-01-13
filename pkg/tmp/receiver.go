package tmp

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	_net "net"
	"os"
	"strconv"

	"github.com/nchaloult/lancp/pkg/input"
	"github.com/nchaloult/lancp/pkg/net"
	"github.com/nchaloult/lancp/pkg/passphrase"
)

// Receive executes appropriate procedures when lancp is run in receive mode. It
// completes an initial passphrase handshake with a sender, creates a
// self-signed TLS certificate and sends it to the sender, establishes a TLS
// connection with the sender, and receives a file.
func Receive() error {
	// TODO: Revisit where these are initialized. Find new homes for these once
	// this big ass func is broken up into pieces.
	portAsStr := fmt.Sprintf(":%d", port)
	tlsPortAsStr := fmt.Sprintf(":%d", port+1)

	log.Println("lancp running in receive mode...")

	generatedPassphrase := passphrase.Generate()

	// Display the generated passphrase for the sender to send.
	log.Printf("Passphrase: %s\n", generatedPassphrase)

	// Listen for a broadcast message from the device running in "send mode."
	udpConn, err := _net.ListenPacket("udp4", portAsStr)
	if err != nil {
		return fmt.Errorf("failed to stand up local UDP packet announcer: %v",
			err)
	}

	// Capture the payload that the sender included in their broadcast message.
	// TODO: Shrink buffer size once you've written passphrase generation logic.
	passphrasePayloadBuf := make([]byte, 1024)
	n, senderAddr, err := udpConn.ReadFrom(passphrasePayloadBuf)
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to read broadcast message from sender: %v",
			err)
	}
	passphrasePayload := string(passphrasePayloadBuf[:n])

	// Compare payload with expected payload.
	if passphrasePayload != generatedPassphrase {
		udpConn.Close()
		return fmt.Errorf("got %q from %s, want %q",
			passphrasePayload, senderAddr.String(), generatedPassphrase)
	}
	log.Printf("got %q from %s, matched expected passphrase",
		passphrasePayload, senderAddr.String())

	// Capture user input for the passphrase the sender is presenting.
	capturer, err := input.NewCapturer("➜", false, os.Stdin, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to create a new Capturer: %v", err)
	}
	userInput, err := capturer.CapturePassphrase()
	if err != nil {
		return err
	}

	// Send response message to sender.
	_, err = udpConn.WriteTo([]byte(userInput), senderAddr)
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to send response message to sender: %v", err)
	}

	// Even though the sender hasn't had time to receive and parse the response
	// UDP datagram we just sent, we can close the PacketConn on our end since
	// UDP is a stateless protocol.
	udpConn.Close()

	// Begin standing up TCP server to exchange cert, and prepare to establish a
	// TLS connection with the sender.

	// TODO: This called func lives in sender.go rn. Move this to some new
	// shared location when you refactor everything.
	localAddr, err := net.GetPreferredOutboundAddr()
	if err != nil {
		return fmt.Errorf("failed to get this device's local IP address: %v",
			err)
	}

	// Generate self-signed TLS cert.
	cert, err := net.GenerateSelfSignedCert(localAddr)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %v", err)
	}

	// Listen for the first part of the TCP handshake from the sender. Send the
	// sender the TLS certificate on that connection.
	tcpLn, err := _net.Listen("tcp", portAsStr)
	if err != nil {
		return fmt.Errorf("failed to start a TCP listener: %v", err)
	}
	// Block until the sender initiates the handshake.
	tcpConn, err := tcpLn.Accept()
	if err != nil {
		return fmt.Errorf("failed to establish TCP connection with sender: %v",
			err)
	}

	// Send TLS certificate to the sender.
	tcpConn.Write(cert.Bytes)

	// Listen for an attempt to establish a TLS connection from the sender.
	tlsCfg, err := net.GetReceiverTLSConfig(cert)
	if err != nil {
		return fmt.Errorf("failed to build TLS config: %v", err)
	}
	tlsLn, err := tls.Listen("tcp", tlsPortAsStr, tlsCfg)
	if err != nil {
		return fmt.Errorf("failed to start a TLS listener: %v", err)
	}
	defer tlsLn.Close()

	// Close up the TCP connection after the TLS listener has already fired up.
	// This is so the sender can try to establish a TLS connection with the
	// receiver immediately after they get the receiver's public key, and not
	// have to make any guesses or assumptions about how long the receiver will
	// take to shut down their TCP listener and spin up their TLS listener.
	tcpConn.Close()
	tcpLn.Close()

	// Block until the sender initiates the handshake.
	tlsConn, err := tlsLn.Accept()
	if err != nil {
		return fmt.Errorf("failed to establish TLS connection with sender: %v",
			err)
	}
	defer tlsConn.Close()

	// Create a file on disk that will eventually store the payload we receive
	// from the sender.
	//
	// TODO: Read the file name that the sender sends first. Right now, the file
	// name is hard-coded by the receiver.
	file, err := os.Create("from-sender")
	if err != nil {
		return fmt.Errorf("failed to create a new file on disk: %v", err)
	}
	defer file.Close()

	// Receive file's bytes from the sender.

	// TODO: Is this an okay size for this buffer? How big could it ever get?
	fileSizeBuf := make([]byte, 10)
	n, err = tlsConn.Read(fileSizeBuf)
	if err != nil {
		return fmt.Errorf("failed to read file size from sender: %v", err)
	}
	fileSizeAsStr := string(fileSizeBuf[:n])
	fileSize, err := strconv.ParseInt(fileSizeAsStr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to convert %q to an int: %v",
			fileSizeAsStr, err)
	}

	// Write that payload to a file on disk.
	var receivedBytes int64
	for {
		if (fileSize - receivedBytes) < filePayloadBufSize {
			io.CopyN(file, tlsConn, (fileSize - receivedBytes))
			tlsConn.Read(make([]byte, (receivedBytes+filePayloadBufSize)-fileSize))
			break
		}

		io.CopyN(file, tlsConn, filePayloadBufSize)
		receivedBytes += filePayloadBufSize
	}

	log.Printf("received %d bytes from sender\n", receivedBytes)

	return nil
}
