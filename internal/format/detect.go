// Package format provides auto-detection and parsing of different data formats.
package format

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/danqzq/rift/internal/stream"
)

// FormatType represents the detected format of input data.
type FormatType string

const (
	FormatJSON FormatType = "json"
	FormatCSV  FormatType = "csv"
	FormatRaw  FormatType = "raw"
)

// ParseResult holds the result of parsing a line of input.
type ParseResult struct {
	Points []stream.DataPoint
	Format FormatType
}

// Detect analyzes a line of input and returns the detected format.
func Detect(line string) FormatType {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return FormatRaw
	}

	if (strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}")) ||
		(strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]")) {
		return FormatJSON
	}

	if strings.Contains(line, ",") || strings.Contains(line, "\t") {
		return FormatCSV
	}

	return FormatRaw
}

// Parse attempts to parse a line using the specified format.
func Parse(line string, format FormatType) ParseResult {
	switch format {
	case FormatJSON:
		return parseJSON(line)
	case FormatCSV:
		return parseCSV(line)
	default:
		return parseRaw(line)
	}
}

// AutoParse detects the format and parses the line automatically.
func AutoParse(line string) ParseResult {
	format := Detect(line)
	result := Parse(line, format)
	result.Format = format
	return result
}

// parseJSON handles JSON objects and arrays.
func parseJSON(line string) ParseResult {
	line = strings.TrimSpace(line)
	result := ParseResult{Format: FormatJSON}

	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err == nil {
		points := extractFromMap(obj, line)
		result.Points = points
		return result
	}

	var arr []any
	if err := json.Unmarshal([]byte(line), &arr); err == nil {
		points := extractFromArray(arr, line)
		result.Points = points
		return result
	}

	return result
}

// extractFromMap extracts DataPoints from a JSON object.
func extractFromMap(obj map[string]any, raw string) []stream.DataPoint {
	var points []stream.DataPoint

	valueFields := []string{"value", "val", "v", "metric", "count", "amount", "num"}
	labelFields := []string{"label", "name", "key", "metric", "type", "service", "field"}

	var value float64
	var label string
	var foundValue bool

	for _, vf := range valueFields {
		if v, ok := obj[vf]; ok {
			if f, ok := toFloat64(v); ok {
				value = f
				foundValue = true
				break
			}
		}
	}

	for _, lf := range labelFields {
		if l, ok := obj[lf]; ok {
			if s, ok := l.(string); ok {
				label = s
				break
			}
		}
	}

	if foundValue {
		dp := stream.NewLabeledDataPoint(label, value)
		dp.Raw = raw
		points = append(points, dp)
		return points
	}

	for k, v := range obj {
		if f, ok := toFloat64(v); ok {
			dp := stream.NewLabeledDataPoint(k, f)
			dp.Raw = raw
			points = append(points, dp)
		}
	}

	return points
}

// extractFromArray extracts DataPoints from a JSON array.
func extractFromArray(arr []any, raw string) []stream.DataPoint {
	var points []stream.DataPoint

	for _, item := range arr {
		switch v := item.(type) {
		case float64:
			dp := stream.NewDataPoint(v)
			dp.Raw = raw
			points = append(points, dp)
		case map[string]any:
			subPoints := extractFromMap(v, raw)
			points = append(points, subPoints...)
		}
	}

	return points
}

// toFloat64 attempts to convert an interface to float64.
func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// parseCSV handles comma or tab separated values.
func parseCSV(line string) ParseResult {
	result := ParseResult{Format: FormatCSV}
	line = strings.TrimSpace(line)

	sep := ","
	if strings.Contains(line, "\t") && !strings.Contains(line, ",") {
		sep = "\t"
	}

	parts := strings.Split(line, sep)
	if len(parts) == 0 {
		return result
	}

	if len(parts) == 2 {
		value, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err == nil {
			label := strings.TrimSpace(parts[0])
			dp := stream.NewLabeledDataPoint(label, value)
			dp.Raw = line
			result.Points = []stream.DataPoint{dp}
			return result
		}
	}

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if value, err := strconv.ParseFloat(part, 64); err == nil {
			dp := stream.NewLabeledDataPoint(strconv.Itoa(i), value)
			dp.Raw = line
			result.Points = append(result.Points, dp)
		}
	}

	return result
}

var numberPattern = regexp.MustCompile(`[-+]?[0-9]*\.?[0-9]+`)

// parseRaw extracts numeric values from arbitrary text.
func parseRaw(line string) ParseResult {
	result := ParseResult{Format: FormatRaw}
	line = strings.TrimSpace(line)

	if len(line) == 0 {
		return result
	}

	if value, err := strconv.ParseFloat(line, 64); err == nil {
		dp := stream.NewDataPoint(value)
		dp.Raw = line
		result.Points = []stream.DataPoint{dp}
		return result
	}

	matches := numberPattern.FindAllString(line, -1)
	for _, match := range matches {
		if value, err := strconv.ParseFloat(match, 64); err == nil {
			dp := stream.NewDataPoint(value)
			dp.Raw = line
			result.Points = append(result.Points, dp)
		}
	}

	return result
}
