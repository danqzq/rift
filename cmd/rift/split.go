package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/danqzq/rift/internal/chart"
	"github.com/danqzq/rift/internal/format"
	"github.com/danqzq/rift/internal/layout"
	"github.com/danqzq/rift/internal/route"
	"github.com/danqzq/rift/internal/stream"
)

// Split command: route a single input to multiple charts.
func runSplit(args []string) error {
	fs := flag.NewFlagSet("split", flag.ExitOnError)
	field := fs.String("field", "label", "field to route on")
	var routes arrayFlags
	fs.Var(&routes, "route", "routing rule: key:charttype (repeatable)")
	fs.Parse(args)

	if len(routes) == 0 {
		return fmt.Errorf("no routes specified, use --route flag")
	}

	ctx, cancel := setupContext()
	defer cancel()

	router := route.NewRouter()
	regions := make([]*layout.Region, 0, len(routes))

	termWidth, termHeight, _ := layout.GetTerminalSize()
	regionHeight := termHeight / len(routes)

	for i, routeSpec := range routes {
		parts := strings.SplitN(routeSpec, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid route spec %q, expected key:charttype", routeSpec)
		}

		key := strings.TrimSpace(parts[0])
		chartType := strings.TrimSpace(parts[1])

		var sel route.Selector
		if *field != "" {
			sel = route.NewFieldSelector(*field, key)
		} else {
			sel = route.ParseSelector(key)
		}

		var c chart.Chart
		config := chart.Config{Label: key}

		switch chartType {
		case "sparkline":
			c = chart.NewSparkline(config)
		case "bar":
			c = chart.NewBar(config)
		case "counter":
			c = chart.NewCounter(config)
		default:
			return fmt.Errorf("unknown chart type %q", chartType)
		}

		w := stream.NewFixedWindow(100)
		router.AddRoute(&route.Route{
			Selector:  sel,
			ChartType: chartType,
			Chart:     c,
			Window:    w,
		})

		region := layout.NewRegion(0, i*regionHeight, termWidth, regionHeight)
		region.Chart = c
		region.Window = w
		region.Label = key
		regions = append(regions, region)
	}

	reader := stream.NewLineReader(ctx, os.Stdin)
	renderer := layout.NewRenderer(regions)

	layout.HideCursor()
	defer layout.ShowCursor()
	renderer.Clear()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case line, ok := <-reader.Lines():
			if !ok {
				renderer.Clear()
				renderer.Render()
				time.Sleep(2 * time.Second)
				return nil
			}

			result := format.AutoParse(line)
			for _, point := range result.Points {
				router.Route(point)
			}

		case <-ticker.C:
			renderer.Clear()
			renderer.Render()

		case err := <-reader.Errors():
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		}
	}
}
