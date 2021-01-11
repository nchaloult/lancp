package input

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Capturer displays input prompts to the user, and captures user input from
// stdin.
type Capturer struct {
	// CaretCharacter is printed to indicate to the user that they should
	// provide input. Ex: >
	CaretCharacter string

	// isForReceiversPassphrase stores whether this Capturer should prompt the
	// user for the passphrase displayed on the receiver's machine. It affects
	// the prompt message that's printed to stdout.
	isForReceiversPassphrase bool

	// inputReader is the Reader interface where user input is read from. Should
	// be os.Stdin in production. Helpful when writing tests.
	inputReader io.Reader

	// promptWriter is the Writer interface where prompts for input are printed/
	// written. Should be os.Stdout in production. Helpful when writing tests.
	promptWriter io.Writer
}

// NewCapturer returns a pointer to a new Capturer struct initialized with the
// provided caret character.
func NewCapturer(
	caretCharacter string, isForReceiversPassphrase bool,
	inputReader io.Reader, promptWriter io.Writer,
) (*Capturer, error) {
	if len(caretCharacter) != 1 && len([]rune(caretCharacter)) != 1 {
		return nil, fmt.Errorf("caretCharacter must be of length 1 (ideally"+
			" a symbol), got: %s", caretCharacter)
	}

	return &Capturer{
		CaretCharacter:           caretCharacter,
		isForReceiversPassphrase: isForReceiversPassphrase,
		inputReader:              inputReader,
		promptWriter:             promptWriter,
	}, nil
}

// CapturePassphrase prompts the user to enter the passphrase that's displayed
// on the other machine running lancp, and returns their input.
func (c *Capturer) CapturePassphrase() (string, error) {
	inputReader := bufio.NewReader(c.inputReader)

	var machineName string
	if c.isForReceiversPassphrase {
		machineName = "receiver"
	} else {
		machineName = "sender"
	}
	// The log pkg doesn't let you print without a newline char at the end.
	fmt.Fprintf(c.promptWriter,
		"Enter the passphrase displayed on the %s's machine:\n%s ",
		machineName, c.CaretCharacter)

	userInput, err := inputReader.ReadString('\n')
	if err != nil {
		// TODO: Should we handle the case where err == io.EOF differently?
		return "", err
	}

	// Convert all CRLF line endings to LF endings.
	// TODO: Is this necessary? Even on Windows machines?
	userInput = strings.Replace(userInput, "\n", "", -1)

	return userInput, nil
}
