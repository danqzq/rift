package chart

import (
	"fmt"
	"time"

	"github.com/danqzq/rift/internal/stream"
)

// Counter renders a single large numeric value.
type Counter struct {
	Config
	ShowRate bool // if true, compute and show events per second
	lastVal  float64
	lastTime time.Time
}

// NewCounter creates a new counter chart.
func NewCounter(config Config) *Counter {
	return &Counter{
		Config:   config,
		ShowRate: false,
	}
}

// Type returns "counter".
func (c *Counter) Type() string {
	return "counter"
}

// Render generates a large numeric display.
func (c *Counter) Render(w *stream.Window, width, height int) string {
	last, ok := w.Last()
	if !ok {
		return "0"
	}

	value := last.Value

	var rate float64
	var rateStr string
	if c.ShowRate {
		now := time.Now()
		if !c.lastTime.IsZero() {
			elapsed := now.Sub(c.lastTime).Seconds()
			if elapsed > 0 {
				rate = (value - c.lastVal) / elapsed
				rateStr = fmt.Sprintf(" (%.1f/s)", rate)
			}
		}
		c.lastVal = value
		c.lastTime = now
	}

	valueStr := fmt.Sprintf("%.2f", value)

	if c.Label != "" {
		return fmt.Sprintf("%s: %s%s", c.Label, valueStr, rateStr)
	}

	return valueStr + rateStr
}
