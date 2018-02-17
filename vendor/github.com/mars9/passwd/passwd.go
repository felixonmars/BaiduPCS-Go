package passwd

import "errors"

var ErrUnsupportedOS = errors.New("not supported")

var passFunc func(prompt string) ([]byte, error)

// Get displays a prompt to, and reads in a password from /dev/tty
// (/dev/cons). Get turns off character echoing while reading the
// password. The calling process should zero the password as soon
// as possible.
func Get(prompt string) ([]byte, error) {
	if passFunc == nil {
		return nil, ErrUnsupportedOS
	}
	return passFunc(prompt)
}
