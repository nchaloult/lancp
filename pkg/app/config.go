package app

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	_net "net"
	"os"

	"github.com/nchaloult/lancp/pkg/net"
	"github.com/nchaloult/lancp/pkg/passphrase"
	"github.com/nchaloult/lancp/pkg/receiver"
)

// TODO: temporary! This config const should be read in from a global config,
// or maybe even provided as a command-line arg.
const (
	passphrasePayloadBufSize  = 32
	defaultFilePayloadBufSize = 8192
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

// Receive executes appropriate procedures when lancp is run in receive mode. It
// completes an initial passphrase handshake with a sender, creates a
// self-signed TLS certificate and sends it to the sender, establishes a TLS
// connection with the sender, and receives a file.
func (c *Config) Receive() error {
	log.Println("lancp running in receive mode...")

	generatedPassphrase := passphrase.Generate()

	// Create a UDP listener for HandshakeConductor to use.
	udpConn, err := _net.ListenPacket("udp4", c.Port)
	if err != nil {
		return fmt.Errorf("failed to stand up local UDP packet announcer: %v",
			err)
	}

	// Have a HandshakeConductor perform the receiver's responsibilities of the
	// lancp handshake.
	localAddr, err := net.GetPreferredOutboundAddr()
	if err != nil {
		return fmt.Errorf("failed to get this device's local IP address: %v",
			err)
	}
	localAddrAsStr := localAddr.String()
	localAddrAsStr += c.Port
	hc, err := net.NewHandshakeConductor(
		udpConn, passphrasePayloadBufSize, generatedPassphrase, localAddrAsStr,
	)
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to create a new HandshakeConductor: %v", err)
	}
	if err = hc.PerformHandshakeAsReceiver(); err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to perform the receiver's responsibilities"+
			" in the lancp handshake: %v", err)
	}

	// Even though the sender hasn't had time to complete it's last piece of the
	// handshake (check the passphrase guess that we sent), we can close the
	// UDP listener on our end since UDP is a stateless protocol.
	//
	// If the sender decides we sent the wrong passphrase, it just won't attempt
	// to establish a TCP connection with us, and we'll time out.
	udpConn.Close()

	// Begin standing up TCP server to exchange cert, and prepare to establish a
	// TLS connection with the sender.

	// Generate self-signed TLS cert.
	cert, err := net.GenerateSelfSignedCert(localAddr)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %v", err)
	}

	// Listen for the first part of the TCP handshake from the sender. Send the
	// sender the TLS certificate on that connection.
	tcpLn, err := _net.Listen("tcp", c.Port)
	if err != nil {
		return fmt.Errorf("failed to start a TCP listener: %v", err)
	}
	// Block until the sender attempts to establish a TCP connection with us.
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
	tlsLn, err := tls.Listen("tcp", c.TLSPort, tlsCfg)
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

	// Receive file's name and size from the sender.

	// TODO: Is this an okay size for this buffer? How big could it ever get?
	// https://www.ibm.com/support/knowledgecenter/SSEQVQ_8.1.10/client/c_cmd_filespecsyntax.html
	fileNameBuf := make([]byte, 1024)
	n, err := tlsConn.Read(fileNameBuf)
	if err != nil {
		return fmt.Errorf("failed to read file name from sender: %v", err)
	}
	fileNameAsStr := string(fileNameBuf[:n])

	// Combo of answers from https://stackoverflow.com/questions/35371385/how-can-i-convert-an-int64-into-a-byte-array-in-go
	fileSizeBuf := make([]byte, binary.MaxVarintLen64)
	_, err = tlsConn.Read(fileSizeBuf)
	if err != nil {
		return fmt.Errorf("failed to read file size from sender: %v", err)
	}
	fileSize, _ := binary.Varint(fileSizeBuf)

	// Create a file on disk that will eventually store the payload we receive
	// from the sender.
	//
	// TODO: Read the file name that the sender sends first. Right now, the file
	// name is hard-coded by the receiver.
	receivedFile, err := os.OpenFile(
		fileNameAsStr,
		os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		// File already exists with the same name as the one the receiver is
		// trying to send. Save under a modified name.
		fileNameAsStr = fmt.Sprintf("lancp-%s", fileNameAsStr)
		receivedFile, err = os.OpenFile(
			fileNameAsStr,
			os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	}
	if err != nil {
		return fmt.Errorf("failed to create a new file on disk: %v", err)
	}
	defer receivedFile.Close()

	// Write that payload to a file on disk.
	_, err = receiver.WritePayloadToFile(
		receivedFile, fileSize, tlsConn,
	)
	if err != nil {
		return fmt.Errorf("failed to write file to disk: %v", err)
	}

	log.Printf("wrote %q to disk\n", fileNameAsStr)
	return nil
}

// Send executes appropriate procedures when lancp is run in send mode. It
// completes an initial passphrase handshake with a receiver, receives a TLS
// certificate from the receiver, establishes a TLS connection with the
// receiver, and sends a file.
func (c *Config) Send() error {
	log.Println("lancp running in send mode...")

	generatedPassphrase := passphrase.Generate()

	// Send broadcast message to find the device running in "receive mode".

	// Get UDP broadcast address.
	localAddr, err := net.GetPreferredOutboundAddr()
	if err != nil {
		return fmt.Errorf("failed to get this device's local IP address: %v",
			err)
	}
	localAddrAsStr := localAddr.String()
	broadcastAddr, err := net.GetBroadcastAddr(localAddrAsStr, c.Port)
	if err != nil {
		return fmt.Errorf("failed to get UDP broadcast address: %v", err)
	}
	broadcastUDPAddr, err := _net.ResolveUDPAddr("udp4", broadcastAddr)
	if err != nil {
		return fmt.Errorf("failed to turn broadcast addr into UDPAddr struct:"+
			" %v", err)
	}

	// Create a UDP listener for HandshakeConductor to use.
	udpConn, err := _net.ListenPacket("udp4", c.Port)
	if err != nil {
		return fmt.Errorf("failed to stand up local UDP packet announcer: %v",
			err)
	}

	// Have a HandshakeConductor perform the sender's responsibilities of the
	// lancp handshake.
	hc, err := net.NewHandshakeConductor(
		udpConn,
		passphrasePayloadBufSize,
		generatedPassphrase,
		localAddrAsStr+c.Port,
	)
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to create a new HandshakeConductor: %v", err)
	}
	receiverAddr, err := hc.PerformHandshakeAsSender(broadcastUDPAddr)
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to perform the sender's responsibilities"+
			" in the lancp handshake: %v", err)
	}

	// At this point, we aren't expecting to get any more UDP datagrams from the
	// receiver. Since UDP is a stateless protocol, we can close the PacketConn
	// on our end.
	udpConn.Close()

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
		receiverIP+c.TLSPort,
		tlsCfg,
	)
	if err != nil {
		return fmt.Errorf("failed to establish TLS conn with receiver: %v", err)
	}
	defer tlsConn.Close()

	// TODO: should we try to look for and open this file on disk before we do
	// any of the logic in this big Send() func? Maybe that could happen after
	// we validate that the command line arg looks like a real path?
	file, err := os.Open(c.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %s: %v", c.FilePath, err)
	}

	// Send file name and size to receiver.
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to retrieve info about the file %s: %v",
			c.FilePath, err)
	}
	fileName := fileInfo.Name()
	tlsConn.Write([]byte(fileName))

	// Combo of answers from https://stackoverflow.com/questions/35371385/how-can-i-convert-an-int64-into-a-byte-array-in-go
	//
	// We're wasting 1-2 bytes of space by making fileSizeBuf large enough to
	// hold a signed 64-bit integer, but I'm fine with that for the sake of
	// convenience :)
	fileSizeBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(fileSizeBuf, fileInfo.Size())
	tlsConn.Write(fileSizeBuf[:n])

	// Send file to the receiver.
	log.Println("sending...")
	filePayloadSize := min(fileInfo.Size(), defaultFilePayloadBufSize)
	filePayloadBuf := make([]byte, filePayloadSize)
	for {
		_, err := file.Read(filePayloadBuf)
		if err == io.EOF {
			break
		}

		tlsConn.Write(filePayloadBuf)
	}

	return nil
}

// This ought to be in the standard library imo.
func min(x, y int64) int64 {
	if x < y {
		return x
	}

	return y
}
