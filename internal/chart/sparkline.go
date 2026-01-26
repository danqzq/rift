package chart

import (
	"fmt"
	"strings"

	"github.com/danqzq/rift/internal/stream"
)

// Sparkline renders a dense time-series using unicode block characters.
type Sparkline struct {
	Config
}

// NewSparkline creates a new sparkline chart.
func NewSparkline(config Config) *Sparkline {
	return &Sparkline{Config: config}
}

// Type returns "sparkline".
func (s *Sparkline) Type() string {
	return "sparkline"
}

// Unicode block characters for 8 levels (0-7).
var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Render generates a sparkline visualization.
func (s *Sparkline) Render(w *stream.Window, width, height int) string {
	points := w.Points()
	if len(points) == 0 {
		return ""
	}

	if len(points) > width {
		points = points[len(points)-width:]
	}
	min, max := s.getScale(w)
	if min == max {
		// All values are the same, render middle block
		return strings.Repeat(string(sparkChars[len(sparkChars)/2]), len(points))
	}

	// Build sparkline
	var sb strings.Builder
	for _, p := range points {
		// Normalize value to 0-1 range
		normalized := (p.Value - min) / (max - min)
		// Map to character index (0-7)
		idx := int(normalized * float64(len(sparkChars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkChars) {
			idx = len(sparkChars) - 1
		}
		sb.WriteRune(sparkChars[idx])
	}

	// Add label if configured
	if s.Label != "" {
		return fmt.Sprintf("%s: %s", s.Label, sb.String())
	}

	return sb.String()
}

// getScale returns the min and max values for scaling.
func (s *Sparkline) getScale(w *stream.Window) (min, max float64) {
	if s.Min != nil {
		min = *s.Min
	} else {
		min, _ = w.Scale()
	}

	if s.Max != nil {
		max = *s.Max
	} else {
		_, max = w.Scale()
	}

	return min, max
}
