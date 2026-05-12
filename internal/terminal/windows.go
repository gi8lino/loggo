//go:build windows

package terminal

import (
	"io"
	"os"
)

// Open opens terminal input and output on Windows.
func Open(_ *os.File, _ io.Writer) (Session, error) {
	closers := []func(){}

	conIn, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		return Session{}, err
	}

	closers = append(closers, func() { _ = conIn.Close() })

	conOut, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		closeAll(closers)

		return Session{}, err
	}

	closers = append(closers, func() { _ = conOut.Close() })

	return Session{
		Input:  conIn,
		Output: conOut,
		Close:  func() { closeAll(closers) },
	}, nil
}
