package layout

import (
	"fmt"
	"strconv"
	"strings"
)

// Grid represents a grid layout specification.
type Grid struct {
	Rows    int
	Cols    int
	Regions []*Region
}

// ParseGrid parses a grid specification like "2x2" or "3x1".
func ParseGrid(spec string) (*Grid, error) {
	parts := strings.Split(spec, "x")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid grid spec %q, expected format: ROWSxCOLS", spec)
	}

	rows, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || rows < 1 {
		return nil, fmt.Errorf("invalid rows in grid spec %q", spec)
	}

	cols, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || cols < 1 {
		return nil, fmt.Errorf("invalid cols in grid spec %q", spec)
	}

	return &Grid{
		Rows:    rows,
		Cols:    cols,
		Regions: make([]*Region, 0, rows*cols),
	}, nil
}

// Calculate divides the terminal into equal regions based on the grid layout.
func (g *Grid) Calculate(termWidth, termHeight int) []*Region {
	regions := make([]*Region, 0, g.Rows*g.Cols)

	cellWidth := termWidth / g.Cols
	cellHeight := termHeight / g.Rows

	for row := 0; row < g.Rows; row++ {
		for col := 0; col < g.Cols; col++ {
			x := col * cellWidth
			y := row * cellHeight
			w := cellWidth
			h := cellHeight

			// Last column takes remaining width
			if col == g.Cols-1 {
				w = termWidth - x
			}

			// Last row takes remaining height
			if row == g.Rows-1 {
				h = termHeight - y
			}

			regions = append(regions, NewRegion(x, y, w, h))
		}
	}

	g.Regions = regions
	return regions
}
