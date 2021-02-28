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
	caretCharacter string

	// machineName stores the name of the machine that Capturer should ask the
	// user to type the passphrase for. machineName is formatted into the prompt
	// message that Capturer displays when asking for user input.
	machineName string

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
	caretCharacter string,
	machineName string,
	inputReader io.Reader,
	promptWriter io.Writer,
) (*Capturer, error) {
	if len(caretCharacter) != 1 && len([]rune(caretCharacter)) != 1 {
		return nil, fmt.Errorf("caretCharacter must be of length 1 (ideally"+
			" a symbol), got: %s", caretCharacter)
	}

	return &Capturer{
		caretCharacter, machineName, inputReader, promptWriter,
	}, nil
}

// CapturePassphrase prompts the user to enter the passphrase that's displayed
// on the other machine running lancp, and returns their input.
func (c *Capturer) CapturePassphrase() (string, error) {
	inputReader := bufio.NewReader(c.inputReader)

	// The log pkg doesn't let you print without a newline char at the end.
	fmt.Fprintf(c.promptWriter,
		"Enter the passphrase displayed on the %s's machine:\n%s ",
		c.machineName, c.caretCharacter)

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
