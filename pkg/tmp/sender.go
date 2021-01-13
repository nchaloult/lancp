package tmp

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	_net "net"
	"os"
	"strconv"

	"github.com/nchaloult/lancp/pkg/input"
	"github.com/nchaloult/lancp/pkg/net"
	"github.com/nchaloult/lancp/pkg/passphrase"
)

// Send executes appropriate procedures when lancp is run in send mode. It
// completes an initial passphrase handshake with a receiver, receives a TLS
// certificate from the receiver, establishes a TLS connection with the
// receiver, and sends a file.
func Send(filePath string) error {
	log.Println("lancp running in send mode...")

	generatedPassphrase := passphrase.Generate()

	// Send broadcast message to find the device running in "receive mode".
	localAddr, err := net.GetPreferredOutboundAddr()
	if err != nil {
		return fmt.Errorf("failed to get this device's local IP address: %v",
			err)
	}
	localAddrAsStr := localAddr.String()
	broadcastAddr, err := net.GetBroadcastAddr(localAddrAsStr, port)
	if err != nil {
		return fmt.Errorf("failed to get UDP broadcast address: %v", err)
	}
	broadcastUDPAddr, err := _net.ResolveUDPAddr("udp4", broadcastAddr)
	if err != nil {
		return fmt.Errorf("failed to turn broadcast addr into UDPAddr struct:"+
			" %v", err)
	}
	// https://github.com/aler9/howto-udp-broadcast-golang
	udpConn, err := _net.ListenPacket("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to stand up local UDP packet announcer: %v",
			err)
	}

	// Capture user input for the passphrase the receiver is presenting.
	capturer, err := input.NewCapturer("➜", true, os.Stdin, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to create a new Capturer: %v", err)
	}
	userInput, err := capturer.CapturePassphrase()
	if err != nil {
		return fmt.Errorf("failed to capture passphrase input from user: %v",
			err)
	}

	_, err = udpConn.WriteTo([]byte(userInput), broadcastUDPAddr)
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to send UDP broadcast message: %v", err)
	}

	// Display the generated passphrase for the receiver to send.
	log.Printf("Passphrase: %s\n", generatedPassphrase)

	// Listen for the response message from the receiver.
	// TODO: Shrink buffer size once you've written passphrase generation logic.
	passphrasePayloadBuf := make([]byte, 1024)
	n, receiverAddr, err := udpConn.ReadFrom(passphrasePayloadBuf)
	// Ignore messages from ourself (like the broadcast message we just sent
	// out).
	if receiverAddr.String() == fmt.Sprintf("%s:%d", localAddrAsStr, port) {
		// Discard our own broadcast message and continue listening for one more
		// message.
		n, receiverAddr, err = udpConn.ReadFrom(passphrasePayloadBuf)
	}
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to read response message from receiver: %v",
			err)
	}
	passphrasePayload := string(passphrasePayloadBuf[:n])

	// At this point, we aren't expecting to get any more UDP datagrams from the
	// receiver. Since UDP is a stateless protocol, we can close the PacketConn
	// on our end.
	udpConn.Close()

	// Compare payload with expected payload.
	if passphrasePayload != generatedPassphrase {
		return fmt.Errorf("got %q from %s, want %q",
			passphrasePayload, receiverAddr.String(), generatedPassphrase)
	}
	log.Printf("got %q from %s, matched expected passphrase",
		passphrasePayload, receiverAddr.String())

	// Get TLS certificate from receiver through an insecure TCP conn.
	tcpConn, err := _net.Dial("tcp", receiverAddr.String())
	if err != nil {
		return fmt.Errorf("failed to establish TCP connection with sender: %v",
			err)
	}
	cert, err := ioutil.ReadAll(tcpConn)
	if err != nil {
		return fmt.Errorf("failed to receive TLS certificate from receiver: %v",
			err)
	}
	tcpConn.Close()

	// Connect to the receiver's TLS conn with that cert.
	tlsCfg := net.GetSenderTLSConfig(cert)
	// TODO: Wow this is a scuffed way to get the receiver's TLS addr.
	receiverIP := receiverAddr.String()[:len(receiverAddr.String())-5]
	tlsConn, err := tls.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", receiverIP, port+1),
		tlsCfg,
	)
	if err != nil {
		return fmt.Errorf("failed to establish TLS conn with receiver: %v", err)
	}
	defer tlsConn.Close()

	// Send file size to receiver.

	// TODO: should we try to look for and open this file on disk before we do
	// any of the logic in this big Send() func? Maybe that could happen after
	// we validate that the command line arg looks like a real path?
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %s: %v", filePath, err)
	}

	// Send file size to receiver.
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to retrieve info about the file %s: %v",
			filePath, err)
	}
	fileSize := strconv.FormatInt(fileInfo.Size(), 10)
	tlsConn.Write([]byte(fileSize))

	// Send file to the receiver.
	filePayloadBuf := make([]byte, filePayloadBufSize)
	for {
		_, err := file.Read(filePayloadBuf)
		if err == io.EOF {
			break
		}

		tlsConn.Write(filePayloadBuf)
	}

	return nil
}