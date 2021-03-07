package net

import (
	"fmt"
	"testing"
)

func TestGetPortAsString(t *testing.T) {
	tests := []struct {
		port        int
		expectedRes string
		expectedErr error
	}{
		// Hard-coding these port numbers in so that if the minValidPort and
		// maxValidPort constants in port.go are changed in the future, these
		// tests will fail.
		{1025, ":1025", nil},
		{65535, ":65535", nil},
		{4242, ":4242", nil},
		{1024, "", fmt.Errorf("port must be be in the range [%d, %d], got: %d",
			1025, 65535, 1024)},
		{65536, "", fmt.Errorf("port must be be in the range [%d, %d], got: %d",
			1025, 65535, 65536)},
		{42, "", fmt.Errorf("port must be be in the range [%d, %d], got: %d",
			1025, 65535, 42)},
		{8675309, "", fmt.Errorf("port must be be in the range [%d, %d], got: %d",
			1025, 65535, 8675309)},
	}

	for _, c := range tests {
		got, err := GetPortAsString(c.port)

		if (err == nil && c.expectedErr != nil) ||
			(err != nil && c.expectedErr == nil) ||
			(c.expectedErr != nil && err.Error() != c.expectedErr.Error()) {
			t.Errorf("unexpected error, got: \"%v\"\nwant: \"%v\"",
				err, c.expectedErr)
		}
		if c.expectedErr != nil && err.Error() == c.expectedErr.Error() {
			// We expected a specific error, and we got that error. We're done
			// with this test case.
			continue
		}

		if got != c.expectedRes {
			t.Errorf("unexpected result, got: %s\nwant: %s", got, c.expectedRes)
		}
	}
}
