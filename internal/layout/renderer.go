package layout

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Renderer handles full-screen terminal rendering.
type Renderer struct {
	regions []*Region
}

// NewRenderer creates a new terminal renderer.
func NewRenderer(regions []*Region) *Renderer {
	return &Renderer{regions: regions}
}

// GetTerminalSize returns the terminal dimensions.
func GetTerminalSize() (width, height int, err error) {
	fd := int(os.Stdout.Fd())
	width, height, err = term.GetSize(fd)
	if err != nil {
		// Fallback to default size
		return 80, 24, nil
	}
	return width, height, nil
}

// Clear clears the terminal screen.
func (r *Renderer) Clear() {
	fmt.Print("\033[2J") // Clear screen
	fmt.Print("\033[H")  // Move cursor to home
}

// Render draws all regions to the terminal.
func (r *Renderer) Render() {
	for _, region := range r.regions {
		if region.Chart == nil || region.Window == nil {
			continue
		}

		// Render the chart
		content := region.Chart.Render(region.Window, region.Width, region.Height)

		// Position cursor and write content
		r.writeAt(region.X, region.Y, content)
	}

	// TODO: improve later with proper cursor management
}

// writeAt positions the cursor and writes content.
func (r *Renderer) writeAt(x, y int, content string) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Move cursor to position: ESC[row;colH
		fmt.Printf("\033[%d;%dH", y+i+1, x+1) // +1 because terminal coords are 1-indexed
		fmt.Print(line)
	}
}

// MoveCursor moves the cursor to the specified position.
func MoveCursor(x, y int) {
	fmt.Printf("\033[%d;%dH", y+1, x+1)
}

// HideCursor hides the terminal cursor.
func HideCursor() {
	fmt.Print("\033[?25l")
}

// ShowCursor shows the terminal cursor.
func ShowCursor() {
	fmt.Print("\033[?25h")
}
