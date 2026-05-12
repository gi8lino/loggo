package terminal

import (
	"io"
	"os"
)

// Session contains terminal input, output, and cleanup.
type Session struct {
	Input  *os.File
	Output io.Writer
	Close  func()
}

// isTerminalFile reports whether file is attached to a terminal.
func isTerminalFile(file *os.File) bool {
	if file == nil {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}

// closeAll closes all collected terminal resources.
func closeAll(closers []func()) {
	for _, closeFn := range closers {
		closeFn()
	}
}

// noopClose is used when no terminal resources need cleanup.
func noopClose() {}
