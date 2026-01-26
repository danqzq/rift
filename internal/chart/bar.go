package chart

import (
	"fmt"
	"sort"
	"strings"

	"github.com/danqzq/rift/internal/stream"
)

// Bar renders horizontal bars for categorical comparisons.
type Bar struct {
	Config
	AutoSort bool // if true, sort by value descending
}

// NewBar creates a new bar chart.
func NewBar(config Config) *Bar {
	return &Bar{
		Config:   config,
		AutoSort: true, // default to sorting
	}
}

// Type returns "bar".
func (b *Bar) Type() string {
	return "bar"
}

// barEntry holds a label and its aggregated value.
type barEntry struct {
	label string
	value float64
	count int
}

// Render generates horizontal bar visualization.
func (b *Bar) Render(w *stream.Window, width, height int) string {
	points := w.Points()
	if len(points) == 0 {
		return ""
	}

	// Aggregate values by label
	entries := b.aggregate(points)

	// Sort if configured
	if b.AutoSort {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].value > entries[j].value
		})
	}

	// Limit to available height
	if len(entries) > height {
		entries = entries[:height]
	}

	// Find max value for scaling
	maxVal := 0.0
	for _, e := range entries {
		if e.value > maxVal {
			maxVal = e.value
		}
	}

	if b.Max != nil {
		maxVal = *b.Max
	}

	// Find longest label for alignment
	maxLabelLen := 0
	for _, e := range entries {
		if len(e.label) > maxLabelLen {
			maxLabelLen = len(e.label)
		}
	}

	// Render bars
	var sb strings.Builder
	barWidth := width - maxLabelLen - 10 // space for label + value
	if barWidth < 1 {
		barWidth = 1
	}

	for _, e := range entries {
		// Calculate bar length
		barLen := 0
		if maxVal > 0 {
			barLen = int((e.value / maxVal) * float64(barWidth))
		}
		if barLen < 0 {
			barLen = 0
		}

		// Render: "label  █████ value"
		label := e.label
		if label == "" {
			label = "?"
		}
		bar := strings.Repeat("█", barLen)
		sb.WriteString(fmt.Sprintf("%-*s %s %.2f\n", maxLabelLen, label, bar, e.value))
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// aggregate combines points by label, computing average for each.
func (b *Bar) aggregate(points []stream.DataPoint) []barEntry {
	m := make(map[string]*barEntry)

	for _, p := range points {
		label := p.Label
		if label == "" {
			label = "value"
		}

		if entry, exists := m[label]; exists {
			entry.value += p.Value
			entry.count++
		} else {
			m[label] = &barEntry{
				label: label,
				value: p.Value,
				count: 1,
			}
		}
	}

	// Convert to slice and compute averages
	result := make([]barEntry, 0, len(m))
	for _, entry := range m {
		entry.value /= float64(entry.count)
		result = append(result, *entry)
	}

	return result
}
