package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

const port = 6969

func send(filePath string) error {
	log.Println("lancp running in send mode...")

	// TODO: Generate passphrase that the receiver will need to present.
	generatedPassphrase := "sender"

	// Send broadcast message to find the device running in "receive mode".
	localAddr, err := getPreferredOutboundAddr()
	if err != nil {
		return fmt.Errorf("failed to get this device's local IP address: %v",
			err)
	}
	localAddrAsStr := localAddr.String()
	broadcastAddr, err := getBroadcastAddr(localAddrAsStr, port)
	if err != nil {
		return fmt.Errorf("failed to get UDP broadcast address: %v", err)
	}
	broadcastUDPAddr, err := net.ResolveUDPAddr("udp4", broadcastAddr)
	if err != nil {
		return fmt.Errorf("failed to turn broadcast addr into UDPAddr struct:"+
			" %v", err)
	}
	// https://github.com/aler9/howto-udp-broadcast-golang
	udpConn, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to stand up local UDP packet announcer: %v",
			err)
	}

	// TODO: Capture user input for the passphrase the receiver is presenting.
	payload := "receiver"

	_, err = udpConn.WriteTo([]byte(payload), broadcastUDPAddr)
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to send UDP broadcast message: %v", err)
	}

	// Listen for the response message from the receiver.
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
	tcpConn, err := net.Dial("tcp", receiverAddr.String())
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
	tlsCfg := getSenderTLSConfig(cert)
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

	// Send a file to the receiver.
	// TODO: Pull file path from command-line args.
	filePayload, err := ioutil.ReadFile("main.go")
	if err != nil {
		return fmt.Errorf("failed to read file on disk: %v", err)
	}
	tlsConn.Write(filePayload)

	return nil
}

// getPreferredOutboundAddr finds this device's preferred outbound IPv4 address
// on its local network. It prepares to send a UDP datagram to Google's DNS, but
// doesn't actually send one.
func getPreferredOutboundAddr() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	preferredOutboundAddr := conn.LocalAddr().(*net.UDPAddr)

	return preferredOutboundAddr.IP, nil
}

// getBroadcastAddr accepts a device's preferred outbound IPv4 address, then
// replaces the host bytes in that address with the broadcast host (all 1s).
// After that, it tacks on the provided port number to the address.
//
// For instance, if this device's preferred local address is 192.168.0.69, then
// getBroadcastAddr will return something like 192.168.0.255:8080.
//
// https://stackoverflow.com/a/37382208
func getBroadcastAddr(
	preferredOutboundAddr string,
	port int,
) (string, error) {
	// Validate input.
	if port <= 1024 || port > 65535 {
		return "", fmt.Errorf("port must be in the range (1024, 65535]."+
			" got: %d", port)
	}

	// Remove host bytes.
	hostBytesIndex := strings.LastIndex(preferredOutboundAddr, ".")
	broadcastAddr := preferredOutboundAddr[:hostBytesIndex]
	// Tack on broadcast host (all 1s) & the port number.
	portAsStr := fmt.Sprintf(":%d", port)
	broadcastAddr += ".255" + portAsStr

	return broadcastAddr, nil
}

// getSenderTLSConfig builds a tls.Config object for the sender to use when
// establishing a TLS connection with the receiver. It adds the public key of
// the certificate authority that the receiver created to the config's
// collection of trusted certificate authorities.
func getSenderTLSConfig(certPEM []byte) *tls.Config {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certPEM)

	return &tls.Config{
		RootCAs: certPool,
	}
}
