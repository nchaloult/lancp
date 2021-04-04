package input

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestCapturePassphrase(t *testing.T) {
	tests := []struct {
		stubbedReader                                         io.Reader
		caretChar, passphraseWant, machineName, promptMsgWant string
	}{
		// Sender with CRLF (Windows) line endings.
		{
			strings.NewReader("foo\r\n"),
			">",
			"foo",
			"sender",
			"Enter the passphrase displayed on the sender's machine:\n> ",
		},
		// Receiver with CRLF (Windows) line endings.
		{
			strings.NewReader("foo\r\n"),
			">",
			"foo",
			"receiver",
			"Enter the passphrase displayed on the receiver's machine:\n> ",
		},
		// Sender with LF (Unix-like) line endings.
		{
			strings.NewReader("foo\n"),
			">",
			"foo",
			"sender",
			"Enter the passphrase displayed on the sender's machine:\n> ",
		},
		// Receiver with LF (Unix-like) line endings.
		{
			strings.NewReader("foo\n"),
			">",
			"foo",
			"receiver",
			"Enter the passphrase displayed on the receiver's machine:\n> ",
		},
	}

	for _, c := range tests {
		var stubbedWriter bytes.Buffer
		capturer, err := NewCapturer(
			c.caretChar,
			c.machineName,
			c.stubbedReader,
			&stubbedWriter,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		passphraseGot, err := capturer.CapturePassphrase()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if passphraseGot != c.passphraseWant {
			t.Errorf("CapturePassphrase() returned an unexpected result,"+
				" got: %q, want: %q", passphraseGot, c.passphraseWant)
		}

		promptMsgGot := stubbedWriter.String()
		promptMsgWant := fmt.Sprintf("Enter the passphrase displayed on the"+
			" %s's machine:\n%s ", c.machineName, c.caretChar)
		if promptMsgGot != promptMsgWant {
			t.Errorf("unexpected prompt message, got: %q, want: %q",
				promptMsgGot, promptMsgWant)
		}
	}
}
