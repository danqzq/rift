package layout

import (
	"github.com/danqzq/rift/internal/chart"
	"github.com/danqzq/rift/internal/stream"
)

// Region represents a rectangular area in the terminal.
type Region struct {
	X      int
	Y      int
	Width  int
	Height int

	Chart  chart.Chart
	Window *stream.Window
	Label  string
}

// NewRegion creates a new region with the specified dimensions.
func NewRegion(x, y, width, height int) *Region {
	return &Region{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}
