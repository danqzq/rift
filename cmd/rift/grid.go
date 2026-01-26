package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/danqzq/rift/internal/layout"
)

// Grid command: compose multiple streams into a grid layout.
func runGrid(args []string) error {
	fs := flag.NewFlagSet("grid", flag.ExitOnError)
	var charts arrayFlags
	fs.Var(&charts, "chart", "chart command to run (repeatable)")
	// Handle grid spec being provided before flags (e.g., "grid 2x2 --chart...")
	var gridSpec string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		gridSpec = args[0]
		args = args[1:]
	}

	fs.Parse(args)

	// If not found before flags, check after
	if gridSpec == "" {
		if fs.NArg() < 1 {
			return fmt.Errorf("grid spec required (e.g., 2x2)")
		}
		gridSpec = fs.Arg(0)
	}

	if len(charts) == 0 {
		return fmt.Errorf("no charts specified, use --chart flag")
	}

	// Parse grid layout
	grid, err := layout.ParseGrid(gridSpec)
	if err != nil {
		return err
	}

	termWidth, termHeight, _ := layout.GetTerminalSize()
	regions := grid.Calculate(termWidth, termHeight)

	if len(charts) > len(regions) {
		return fmt.Errorf("too many charts (%d) for grid %s (%d cells)", len(charts), gridSpec, len(regions))
	}

	ctx, cancel := setupContext()
	defer cancel()

	// Start each chart command as a subprocess
	cmds := make([]*exec.Cmd, len(charts))
	outputs := make([]string, len(charts))

	for i, chartCmd := range charts {
		// Execute through shell to support pipes, loops, etc.
		cmd := exec.CommandContext(ctx, "sh", "-c", chartCmd)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Chart %d error: %v\n", i, err)
			outputs[i] = fmt.Sprintf("Error: %v", err)
		} else {
			outputs[i] = string(output)
		}
		cmds[i] = cmd
	}

	// Render grid with outputs
	layout.HideCursor()
	defer layout.ShowCursor()

	// Clear screen
	fmt.Print("\033[2J\033[H")

	// Render each region
	for i, region := range regions {
		if i >= len(outputs) {
			break
		}

		// Position cursor and write output
		lines := strings.Split(strings.TrimSpace(outputs[i]), "\n")
		for j, line := range lines {
			if j >= region.Height {
				break
			}
			// Move cursor: ESC[row;colH
			fmt.Printf("\033[%d;%dH", region.Y+j+1, region.X+1)
			// Truncate line if too wide
			if len(line) > region.Width {
				line = line[:region.Width]
			}
			fmt.Print(line)
		}
	}

	// Move cursor to bottom
	fmt.Printf("\033[%d;1H", termHeight)

	// Wait a bit before exiting
	time.Sleep(100 * time.Millisecond)

	return nil
}
