package ingest

import (
	"bufio"
	"context"
	"io"
	"regexp"
	"strings"
	"time"
)

// Options controls how input lines are batched.
type Options struct {
	BatchSize     int
	FlushInterval time.Duration
	ScannerBuffer int
	ScannerMax    int
	JoinMultiline bool
	MaxEventLines int
}

// Batch contains a group of streamed input lines or a terminal stream event.
type Batch struct {
	Lines []string
	Err   error
	EOF   bool
}

// Start starts reading lines from r and returns batched input messages.
func Start(ctx context.Context, r io.Reader, opts Options) <-chan Batch {
	out := make(chan Batch, 16)
	lines := make(chan string, opts.BatchSize*2)
	errs := make(chan error, 1)

	go scanLines(ctx, r, opts, lines, errs)
	go runBatcher(ctx, opts, lines, errs, out)

	return out
}

// scanLines scans input lines and sends them to the batcher.
func scanLines(ctx context.Context, r io.Reader, opts Options, lines chan<- string, errs chan<- error) {
	defer close(lines)

	scannerBuffer := opts.ScannerBuffer
	if scannerBuffer <= 0 {
		scannerBuffer = 64 * 1024
	}

	scannerMax := opts.ScannerMax
	if scannerMax <= 0 {
		scannerMax = 4 * 1024 * 1024
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, scannerBuffer), scannerMax)

	current := []string{}
	stacktraceMode := false
	maxEventLines := opts.MaxEventLines
	if maxEventLines <= 0 {
		maxEventLines = 256
	}

	for scanner.Scan() {
		line := scanner.Text()

		if !opts.JoinMultiline {
			select {
			case <-ctx.Done():
				return
			case lines <- line:
			}

			continue
		}

		if len(current) == 0 {
			current = append(current, line)
			stacktraceMode = false
			continue
		}

		if shouldContinueEvent(line, stacktraceMode) && len(current) < maxEventLines {
			current = append(current, line)
			stacktraceMode = stacktraceMode || isJavaStacktraceLine(line)
			continue
		}

		select {
		case <-ctx.Done():
			return
		case lines <- strings.Join(current, "\n"):
		}

		current = []string{line}
		stacktraceMode = false
	}

	if len(current) > 0 {
		select {
		case <-ctx.Done():
			return
		case lines <- strings.Join(current, "\n"):
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case <-ctx.Done():
		case errs <- err:
		}
	}
}

var javaExceptionPattern = regexp.MustCompile(`^(?:[A-Za-z_$][A-Za-z0-9_$]*\.)*[A-Za-z_$][A-Za-z0-9_$]*(?:Exception|Error|Throwable|Failure)(?::|$)`)

// shouldContinueEvent reports whether a line should be merged into the current log event.
func shouldContinueEvent(line string, stacktraceMode bool) bool {
	if isJavaStacktraceLine(line) {
		return true
	}

	if !stacktraceMode {
		return false
	}

	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
}

// isJavaStacktraceLine reports whether a line looks like part of a Java traceback.
func isJavaStacktraceLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	switch {
	case strings.HasPrefix(trimmed, "at "):
		return true
	case strings.HasPrefix(trimmed, "Caused by:"):
		return true
	case strings.HasPrefix(trimmed, "Suppressed:"):
		return true
	case strings.HasPrefix(trimmed, "... ") && strings.HasSuffix(trimmed, " more"):
		return true
	case javaExceptionPattern.MatchString(trimmed):
		return true
	default:
		return false
	}
}

// runBatcher groups input lines by size or flush interval.
func runBatcher(ctx context.Context, opts Options, lines <-chan string, errs <-chan error, out chan<- Batch) {
	defer close(out)

	ticker := time.NewTicker(opts.FlushInterval)
	defer ticker.Stop()

	batch := make([]string, 0, opts.BatchSize)

	for {
		select {
		case <-ctx.Done():
			return

		case line, ok := <-lines:
			if !ok {
				var flushed bool
				batch, flushed = flushBatch(ctx, out, batch)
				if !flushed {
					return
				}

				select {
				case err := <-errs:
					sendBatch(ctx, out, Batch{Err: err})
				default:
				}

				sendBatch(ctx, out, Batch{EOF: true})

				return
			}

			batch = append(batch, line)

			if len(batch) >= opts.BatchSize {
				var flushed bool
				batch, flushed = flushBatch(ctx, out, batch)
				if !flushed {
					return
				}
			}

		case <-ticker.C:
			var flushed bool
			batch, flushed = flushBatch(ctx, out, batch)
			if !flushed {
				return
			}
		}
	}
}

// flushBatch sends the current batch when it contains at least one line.
func flushBatch(ctx context.Context, out chan<- Batch, batch []string) ([]string, bool) {
	if len(batch) == 0 {
		return batch[:0], true
	}

	lines := append([]string(nil), batch...)

	if !sendBatch(ctx, out, Batch{Lines: lines}) {
		return batch, false
	}

	return batch[:0], true
}

// sendBatch sends one batch unless the context is canceled.
func sendBatch(ctx context.Context, out chan<- Batch, batch Batch) bool {
	select {
	case <-ctx.Done():
		return false
	case out <- batch:
		return true
	}
}
