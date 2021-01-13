package net

import (
	"fmt"
	_net "net"
	"strings"
)

// GetPreferredOutboundAddr finds this device's preferred outbound IPv4 address
// on its local network. It prepares to send a UDP datagram to Google's DNS, but
// doesn't actually send one.
func GetPreferredOutboundAddr() (_net.IP, error) {
	conn, err := _net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	preferredOutboundAddr := conn.LocalAddr().(*_net.UDPAddr)

	return preferredOutboundAddr.IP, nil
}

// GetBroadcastAddr accepts a device's preferred outbound IPv4 address, then
// replaces the host bytes in that address with the broadcast host (all 1s).
// After that, it tacks on the provided port number to the address.
//
// For instance, if this device's preferred local address is 192.168.0.69, then
// getBroadcastAddr will return something like 192.168.0.255:8080.
//
// https://stackoverflow.com/a/37382208
func GetBroadcastAddr(
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
