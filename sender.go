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

	broadcastAddr, err := getBroadcastAddr(port)
	if err != nil {
		return err
	}
	log.Printf("sending UDP broadcast msg to: %s\n", broadcastAddr)

	return nil
}

// getBroadcastAddr finds this device's preferred outbound IPv4 address, then
// replaces the host bytes in that address with the broadcast host (all 1s).
// After that, it tacks on the provided port number to the address.
//
// For instance, if this device's local address is 192.168.0.69, then
// getBroadcastAddr will return something like 192.168.0.255:8080.
//
// https://stackoverflow.com/a/37382208
func getBroadcastAddr(port int) (string, error) {
	// Validate input.
	if port <= 1024 || port > 65535 {
		return "", fmt.Errorf("port must be in the range (1024, 65535]."+
			" got: %d", port)
	}

	// Get preferred outbound IP address.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	preferredOutboundAddr := conn.LocalAddr().(*net.UDPAddr)

	broadcastAddr := preferredOutboundAddr.IP.String()
	// Remove host bytes.
	hostBytesIndex := strings.LastIndex(broadcastAddr, ".")
	broadcastAddr = broadcastAddr[:hostBytesIndex]
	// Tack on broadcast host (all 1s) & the port number.
	portAsStr := fmt.Sprintf(":%d", port)
	broadcastAddr += ".255" + portAsStr

	return broadcastAddr, nil
}
