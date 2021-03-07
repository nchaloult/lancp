package net

import (
	"fmt"
	_net "net"
	"strings"
	"testing"
)

func TestGetUDPBroadcastAddr(t *testing.T) {
	preferredOutboundAddr, err := independentGetPreferredOutboundAddr()
	if err != nil {
		t.Fatalf("failed to get preferred outbound address before starting"+
			" tests: %v", err)
	}

	// Not testing any ports outside of the expected range. The
	// getLocalListeningAddr func should get its port parameter from the
	// port.GetPortAsString func, ideally, which already tests this.
	tests := []struct {
		port             string
		expectedResAsStr string
	}{
		{
			":1025",
			fmt.Sprintf("%s%s",
				preferredOutboundAddr[:strings.LastIndex(preferredOutboundAddr,
					".")],
				".255:1025"),
		},
		{
			":65535",
			fmt.Sprintf("%s%s",
				preferredOutboundAddr[:strings.LastIndex(preferredOutboundAddr,
					".")],
				".255:65535"),
		},
	}

	for _, c := range tests {
		got, err := GetUDPBroadcastAddr(c.port)
		if err != nil {
			t.Fatalf("unexpected error, got: \"%v\"", err)
		}

		gotAsStr := got.String()
		if gotAsStr != c.expectedResAsStr {
			t.Errorf("unexpected result, got: %q, want: %q",
				gotAsStr, c.expectedResAsStr)
		}
	}
}

func TestGetLocalListeningAddr(t *testing.T) {
	preferredOutboundAddr, err := independentGetPreferredOutboundAddr()
	if err != nil {
		t.Fatalf("failed to get preferred outbound address before starting"+
			" tests: %v", err)
	}

	// Not testing any ports outside of the expected range. The
	// getLocalListeningAddr func should get its port parameter from the
	// port.GetPortAsString func, ideally, which already tests this.
	tests := []struct {
		port        string
		expectedRes string
	}{
		{":1025", fmt.Sprintf("%s:1025", preferredOutboundAddr)},
		{":65535", fmt.Sprintf("%s:65535", preferredOutboundAddr)},
	}

	for _, c := range tests {
		got, err := getLocalListeningAddr(c.port)
		if err != nil {
			t.Fatalf("unexpected error, got: \"%v\"", err)
		}

		if got != c.expectedRes {
			t.Errorf("unexpected result, got: %q, want: %q", got, c.expectedRes)
		}
	}
}

func TestGetBroadcastAddr(t *testing.T) {
	// Not testing any ports outside of the expected range. The
	// getLocalListeningAddr func should get its port parameter from the
	// port.GetPortAsString func, ideally, which already tests this.
	tests := []struct {
		preferredOutboundAddr string
		port                  string
		expectedRes           string
	}{
		{"192.168.0.1", ":1025", "192.168.0.255:1025"},
		{"10.0.0.1", ":1025", "10.0.0.255:1025"},
		{"192.168.0.1", ":65535", "192.168.0.255:65535"},
		{"10.0.0.1", ":65535", "10.0.0.255:65535"},
	}

	for _, c := range tests {
		got := getBroadcastAddr(c.preferredOutboundAddr, c.port)
		if got != c.expectedRes {
			t.Errorf("unexpected result, got: %q, want: %q", got, c.expectedRes)
		}
	}
}

func TestGetTLSAddress(t *testing.T) {
	// Not testing any ports outside of the expected range. The
	// getLocalListeningAddr func should get its port parameter from the
	// port.GetPortAsString func, ideally, which already tests this.
	tests := []struct {
		preferredOutboundAddr string
		port                  string
		expectedRes           string
	}{
		{"192.168.0.1:1025", ":1026", "192.168.0.1:1026"},
		{"10.0.0.1:1025", ":1026", "10.0.0.1:1026"},
		{"192.168.0.1:65535", ":65534", "192.168.0.1:65534"},
		{"10.0.0.1:65535", ":65534", "10.0.0.1:65534"},
	}

	for _, c := range tests {
		got := GetTLSAddress(c.preferredOutboundAddr, c.port)
		if got != c.expectedRes {
			t.Errorf("unexpected result, got: %q, want: %q", got, c.expectedRes)
		}
	}
}

// independentGetPreferredOutboundAddr gets the test runner device's preferred
// outbound address.
//
// We're reimplementing GetPreferredOutboundAddress()'s functionality here. If
// that func stops working in the future, or works differently than we're
// expecting it to, we want the tests in this file to reflect that.
func independentGetPreferredOutboundAddr() (string, error) {
	conn, err := _net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	return conn.LocalAddr().(*_net.UDPAddr).IP.String(), nil
}
