//go:build windows

package terminal

import (
	"io"
	"os"
)

// Open opens terminal input and output on Windows.
func Open(stdIn *os.File, stdErr io.Writer) (Session, error) {
	closers := []func(){}

	input := stdIn
	if !isTerminalFile(stdIn) {
		conIn, err := os.OpenFile("CONIN$", os.O_RDONLY, 0)
		if err != nil {
			return Session{}, err
		}

		input = conIn
		closers = append(closers, func() { _ = conIn.Close() })
	}

	output := stdErr
	if file, ok := stdErr.(*os.File); !ok || !isTerminalFile(file) {
		conOut, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
		if err != nil {
			closeAll(closers)

			return Session{}, err
		}

		output = conOut
		closers = append(closers, func() { _ = conOut.Close() })
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
