package stream

import "io"

// StreamSource defines the interface for reading data points from a stream.
type StreamSource interface {
	// Read returns the next DataPoint from the stream (EOF when done)
	Read() (DataPoint, error)

	// Type returns the detected format of the stream ("json", "csv", "raw").
	Type() string

	// Close releases any resources held by the source.
	Close() error
}

// Ensure io.EOF is used for end-of-stream signaling.
var _ = io.EOF
