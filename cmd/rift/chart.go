package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/danqzq/rift/internal/chart"
	"github.com/danqzq/rift/internal/format"
	"github.com/danqzq/rift/internal/stream"
)

// Bar command: render all input as a single bar chart.
func runBar(_ []string) error {
	ctx, cancel := setupContext()
	defer cancel()

	reader := stream.NewLineReader(ctx, os.Stdin)
	window := stream.NewFixedWindow(1000)

	// Read all input
	for {
		select {
		case <-ctx.Done():
			return nil
		case line, ok := <-reader.Lines():
			if !ok {
				// EOF - render the bar chart
				barChart := chart.NewBar(chart.Config{})
				output := barChart.Render(window, 80, 50)
				fmt.Println(output)
				return nil
			}

			result := format.AutoParse(line)
			for _, point := range result.Points {
				window.Add(point)
			}

		case err := <-reader.Errors():
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
		}
	}
}

// Sparkline command: render all input as a sparkline.
func runSparkline(args []string) error {
	ctx, cancel := setupContext()
	defer cancel()

	reader := stream.NewLineReader(ctx, os.Stdin)
	window := stream.NewFixedWindow(1000)

	var min *float64
	var max *float64
	if len(args) > 0 {
		minVal, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid min value: %w", err)
		}
		min = &minVal
	}
	if len(args) > 1 {
		maxVal, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid max value: %w", err)
		}
		max = &maxVal
	}

	// Read all input
	for {
		select {
		case <-ctx.Done():
			return nil
		case line, ok := <-reader.Lines():
			if !ok {
				// EOF - render the sparkline
				sparklineChart := chart.NewSparkline(chart.Config{
					Min: min,
					Max: max,
				})
				output := sparklineChart.Render(window, 200, 1)
				fmt.Println(output)
				time.Sleep(100 * time.Millisecond)
				return nil
			}

			result := format.AutoParse(line)
			for _, point := range result.Points {
				window.Add(point)
			}

		case err := <-reader.Errors():
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
		}
	}
}
