package chart

import (
	"strings"
	"testing"

	"github.com/danqzq/rift/internal/stream"
)

func TestSparkline_Render(t *testing.T) {
	tests := []struct {
		name    string
		values  []float64
		width   int
		wantLen int
	}{
		{
			name:    "empty window",
			values:  []float64{},
			width:   10,
			wantLen: 0,
		},
		{
			name:    "single value",
			values:  []float64{5},
			width:   10,
			wantLen: 1,
		},
		{
			name:    "ascending values",
			values:  []float64{1, 2, 3, 4, 5},
			width:   10,
			wantLen: 5,
		},
		{
			name:    "more values than width",
			values:  []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			width:   5,
			wantLen: 5, // should only show last 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := stream.NewFixedWindow(100)
			for _, v := range tt.values {
				w.Add(stream.NewDataPoint(v))
			}

			s := NewSparkline(Config{})
			result := s.Render(w, tt.width, 1)

			// Count unicode chars (not bytes)
			charCount := len([]rune(result))
			if charCount != tt.wantLen {
				t.Errorf("got %d chars, want %d chars", charCount, tt.wantLen)
			}
		})
	}
}

func TestSparkline_Type(t *testing.T) {
	s := NewSparkline(Config{})
	if s.Type() != "sparkline" {
		t.Errorf("Type() = %q, want %q", s.Type(), "sparkline")
	}
}

func TestBar_Render(t *testing.T) {
	w := stream.NewFixedWindow(100)
	w.Add(stream.NewLabeledDataPoint("cpu", 45))
	w.Add(stream.NewLabeledDataPoint("memory", 80))
	w.Add(stream.NewLabeledDataPoint("cpu", 55))

	b := NewBar(Config{})
	result := b.Render(w, 50, 10)

	// Should have two lines (cpu and memory)
	lines := strings.Split(result, "\n")
	if len(lines) != 2 {
		t.Errorf("got %d lines, want 2", len(lines))
	}

	// First line should be memory (higher value, auto-sorted)
	if !strings.Contains(lines[0], "memory") {
		t.Errorf("first line should be memory, got: %s", lines[0])
	}
}

func TestCounter_Render(t *testing.T) {
	w := stream.NewFixedWindow(10)
	w.Add(stream.NewDataPoint(42.5))

	c := NewCounter(Config{Label: "Total"})
	result := c.Render(w, 50, 5)

	if !strings.Contains(result, "42.50") {
		t.Errorf("result should contain 42.50, got: %s", result)
	}

	if !strings.Contains(result, "Total") {
		t.Errorf("result should contain label, got: %s", result)
	}
}
