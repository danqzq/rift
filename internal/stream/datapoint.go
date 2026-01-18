// Package stream provides core streaming data types and utilities.
package stream

import "time"

// DataPoint represents a single data observation from a stream.
type DataPoint struct {
	// Timestamp when the data point was received or created.
	Timestamp time.Time

	// Value holds the numeric value of the data point.
	Value float64

	// Label is an optional identifier for categorical data (e.g., "cpu", "memory").
	Label string

	// Raw stores the original input string for debugging purposes.
	Raw string
}

func NewDataPoint(value float64) DataPoint {
	return DataPoint{
		Timestamp: time.Now(),
		Value:     value,
	}
}

func NewLabeledDataPoint(label string, value float64) DataPoint {
	return DataPoint{
		Timestamp: time.Now(),
		Value:     value,
		Label:     label,
	}
}
