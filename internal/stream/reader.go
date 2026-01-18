package stream

import (
	"bufio"
	"context"
	"io"
)

// LineReader provides non-blocking line-by-line reading from an io.Reader.
type LineReader struct {
	lines  chan string
	errors chan error
	done   chan struct{}
	cancel context.CancelFunc
}

func NewLineReader(ctx context.Context, r io.Reader) *LineReader {
	ctx, cancel := context.WithCancel(ctx)

	lr := &LineReader{
		lines:  make(chan string, 100), // buffer to handle bursts
		errors: make(chan error, 1),
		done:   make(chan struct{}),
		cancel: cancel,
	}

	go lr.readLoop(ctx, r)
	return lr
}

// readLoop continuously reads lines and sends them to the channel.
func (lr *LineReader) readLoop(ctx context.Context, r io.Reader) {
	defer close(lr.done)
	defer close(lr.lines)
	defer close(lr.errors)

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		case lr.lines <- scanner.Text():
			// line sent successfully
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case lr.errors <- err:
		default:
			// error channel full, drop the error
		}
	}
}

func (lr *LineReader) Lines() <-chan string {
	return lr.lines
}

func (lr *LineReader) Errors() <-chan error {
	return lr.errors
}

func (lr *LineReader) Done() <-chan struct{} {
	return lr.done
}

func (lr *LineReader) Stop() {
	lr.cancel()
	<-lr.done
}
