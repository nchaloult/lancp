package main

import (
	"fmt"
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
	broadcastAddr, err := getBroadcastAddr(localAddr, port)
	if err != nil {
		return fmt.Errorf("failed to get UDP broadcast address: %v", err)
	}
	broadcastUDPAddr, err := net.ResolveUDPAddr("udp4", broadcastAddr)
	if err != nil {
		return fmt.Errorf("failed to turn broadcast addr into UDPAddr struct:"+
			" %v", err)
	}
	// https://github.com/aler9/howto-udp-broadcast-golang
	conn, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to stand up local UDP packet announcer: %v",
			err)
	}
	// TODO: might wanna do this sooner; don't defer it until the end of this
	// big ass func. Putting this here will make more sense once the logic in
	// this func is split up.
	defer conn.Close()

	// TODO: Capture user input for the passphrase the receiver is presenting.
	payload := "receiver"

	_, err = conn.WriteTo([]byte(payload), broadcastUDPAddr)
	if err != nil {
		return fmt.Errorf("failed to send UDP broadcast message: %v", err)
	}

	// Listen for the response message from the receiver.
	receiverPayloadBuf := make([]byte, 1024)
	n, receiverAddr, err := conn.ReadFrom(receiverPayloadBuf)
	// Ignore messages from ourself (like the broadcast message we just sent
	// out).
	if receiverAddr.String() == fmt.Sprintf("%s:%d", localAddr, port) {
		// Discard our own broadcast message and continue listening for one more
		// message.
		n, receiverAddr, err = conn.ReadFrom(receiverPayloadBuf)
	}
	if err != nil {
		return fmt.Errorf("failed to read response message from receiver: %v",
			err)
	}
	receiverPayload := string(receiverPayloadBuf[:n])

	// Compare payload with expected payload.
	if receiverPayload != generatedPassphrase {
		return fmt.Errorf("got %q from %s, want %q",
			receiverPayload, receiverAddr.String(), generatedPassphrase)
	}
	log.Printf("got %q from %s, matched expected passphrase",
		payload, receiverAddr.String())

	log.Println("At this point, receiver has already started a listener, and" +
		" the sender gets the receiver's cert")

	return nil
}

// getPreferredOutboundAddr finds this device's preferred outbound IPv4 address
// on its local network. It prepares to send a UDP datagram to Google's DNS, but
// doesn't actually send one.
func getPreferredOutboundAddr() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	preferredOutboundAddr := conn.LocalAddr().(*net.UDPAddr)

	return preferredOutboundAddr.IP.String(), nil
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
