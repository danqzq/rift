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
