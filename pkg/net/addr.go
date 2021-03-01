package net

import (
	"fmt"
	_net "net"
	"strings"
)

// GetUDPBroadcastAddr builds a UDP address for the local network that this
// machine is connected to. When a UDP message is sent to this address, every
// device on the same local network will receive it (including the device who
// originally sent the message!), and can read its message if they are listening
// for these messages on the right port.
// https://en.wikipedia.org/wiki/Broadcast_address
//
// It finds this device's preferred outbound IPv4 address, converts that into a
// UDP broadcast address, and tacks on the provided port.
//
// port needs to look like a port string (i.e., ":xxxx" or ":xxxxx").
func GetUDPBroadcastAddr(port string) (*_net.UDPAddr, error) {
	localAddr, err := getPreferredOutboundAddr()
	if err != nil {
		return nil, fmt.Errorf("failed to get this device's local IP address:"+
			" %v", err)
	}
	broadcastAddr := getBroadcastAddr(localAddr.String(), port)
	broadcastUDPAddr, err := _net.ResolveUDPAddr("udp4", broadcastAddr)
	if err != nil {
		return nil, err
	}

	return broadcastUDPAddr, nil
}

// getLocalListeningAddress gets the local address of this machine and appends
// the provided port string to it. Useful for matching a message's return
// address with the local address of a machine to detect loopback messages.
//
// port needs to look like a port string (i.e., ":xxxx" or ":xxxxx").
func getLocalListeningAddr(port string) (string, error) {
	localAddr, err := getPreferredOutboundAddr()
	if err != nil {
		return "", fmt.Errorf("failed to get this device's local IP address:"+
			" %v", err)
	}

	return localAddr.String() + port, nil
}

// getPreferredOutboundAddr finds this device's preferred outbound IPv4 address
// on its local network. It prepares to send a UDP datagram to Google's DNS, but
// doesn't actually send one.
func getPreferredOutboundAddr() (_net.IP, error) {
	conn, err := _net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	preferredOutboundAddr := conn.LocalAddr().(*_net.UDPAddr)

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
	port string,
) string {
	// Remove host bytes.
	hostBytesIndex := strings.LastIndex(preferredOutboundAddr, ".")
	broadcastAddr := preferredOutboundAddr[:hostBytesIndex]
	// Tack on broadcast host (all 1s) & the port number.
	broadcastAddr += ".255" + port

	return broadcastAddr
}

// GetTLSAddress builds an address from a machine's IP and TLS port. It strips
// off the port number from the provided address, and tacks on the provided
// port in its place.
//
// addr must be an IPv4 address. Assumed to already have a port number on it.
//
// port needs to look like a port string (i.e., ":xxxx" or ":xxxxx").
func GetTLSAddress(addr string, port string) string {
	addr = addr[:strings.LastIndex(addr, ":")]
	return addr + port
}
