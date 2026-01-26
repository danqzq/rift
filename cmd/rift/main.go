package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/danqzq/rift/internal/format"
	"github.com/danqzq/rift/internal/stream"
)

func main() {
	// Check for subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "split":
			if err := runSplit(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "grid":
			if err := runGrid(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "bar":
			if err := runBar(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "sparkline":
			if err := runSparkline(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "-h", "--help", "help":
			printHelp()
			return
		}
	}

	// Default behavior: simple display mode
	runSimpleMode()
}

func runSimpleMode() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nShutting down...")
		cancel()
	}()

	reader := stream.NewLineReader(ctx, os.Stdin)
	window := stream.NewFixedWindow(100)

	fmt.Fprintln(os.Stderr, "rift: Waiting for input... (Ctrl+C to exit)")

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-reader.Lines():
			if !ok {
				printSummary(window)
				return
			}

			result := format.AutoParse(line)

			for _, point := range result.Points {
				window.Add(point)
				displayPoint(point, result.Format)
			}

		case err := <-reader.Errors():
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			}
		}
	}
}

func displayPoint(p stream.DataPoint, f format.FormatType) {
	if p.Label != "" {
		fmt.Printf("[%s] %s: %.2f\n", f, p.Label, p.Value)
	} else {
		fmt.Printf("[%s] %.2f\n", f, p.Value)
	}
}

func printSummary(w *stream.Window) {
	if w.Len() == 0 {
		fmt.Fprintln(os.Stderr, "\nNo data points received.")
		return
	}

	min, max := w.Scale()
	fmt.Fprintf(os.Stderr, "\n--- Summary ---\n")
	fmt.Fprintf(os.Stderr, "Points: %d\n", w.Len())
	fmt.Fprintf(os.Stderr, "Min: %.2f\n", min)
	fmt.Fprintf(os.Stderr, "Max: %.2f\n", max)
}

func printHelp() {
	fmt.Println(`rift - Real-time metrics compositor

USAGE:
    rift [COMMAND]

COMMANDS:
    bar          Render input as a bar chart
    sparkline    Render input as a sparkline
    split        Route single input stream to multiple charts
    grid         Compose multiple streams into a grid layout
    help         Show this message

Run 'rift split -h' or 'rift grid -h' for command-specific help.

When run without commands, rift reads from stdin and displays parsed values.`)
}
