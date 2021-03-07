package net

import (
	"fmt"
	_net "net"
	"testing"
)

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
