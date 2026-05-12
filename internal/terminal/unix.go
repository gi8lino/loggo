//go:build !windows

package terminal

import (
	"io"
	"os"
)

// Open opens terminal input and output on Unix-like systems.
func Open(stdIn *os.File, stdErr io.Writer) (Session, error) {
	closers := []func(){}

	input := stdIn
	if !isTerminalFile(stdIn) {
		tty, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
		if err != nil {
			return Session{}, err
		}

		input = tty
		closers = append(closers, func() { _ = tty.Close() })
	}

	output := stdErr
	if file, ok := stdErr.(*os.File); !ok || !isTerminalFile(file) {
		tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
		if err != nil {
			closeAll(closers)

			return Session{}, err
		}

		output = tty
		closers = append(closers, func() { _ = tty.Close() })
	}

	closeFn := noopClose
	if len(closers) > 0 {
		closeFn = func() { closeAll(closers) }
	}

	return Session{
		Input:  input,
		Output: output,
		Close:  closeFn,
	}, nil
}
