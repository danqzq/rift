package chart

import "github.com/danqzq/rift/internal/stream"

// Chart defines the interface for rendering visualizations.
type Chart interface {
	// Render generates the visual representation as a string.
	// width and height specify the available space in terminal cells.
	Render(w *stream.Window, width, height int) string

	// Type returns the chart type identifier.
	Type() string
}

// Config holds common chart configuration options.
type Config struct {
	Label string
	Min   *float64 // nil means auto-scale
	Max   *float64 // nil means auto-scale
	Color string   // "red", "green", "yellow", etc.
}
