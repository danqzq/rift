package route

import (
	"testing"

	"github.com/danqzq/rift/internal/stream"
)

func TestFieldSelector_Matches(t *testing.T) {
	tests := []struct {
		name     string
		selector *FieldSelector
		point    stream.DataPoint
		want     bool
	}{
		{
			name:     "matches label",
			selector: NewFieldSelector("label", "cpu"),
			point:    stream.NewLabeledDataPoint("cpu", 45),
			want:     true,
		},
		{
			name:     "does not match label",
			selector: NewFieldSelector("label", "memory"),
			point:    stream.NewLabeledDataPoint("cpu", 45),
			want:     false,
		},
		{
			name:     "matches metric field",
			selector: NewFieldSelector("metric", "latency"),
			point:    stream.NewLabeledDataPoint("latency", 23.5),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.selector.Matches(tt.point)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlwaysSelector_Matches(t *testing.T) {
	sel := &AlwaysSelector{}
	p := stream.NewDataPoint(42)

	if !sel.Matches(p) {
		t.Error("AlwaysSelector should always match")
	}
}

func TestParseSelector(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string // type name
	}{
		{"empty", "", "*"},
		{"star", "*", "*"},
		{"field equals", "metric=cpu", "metric=cpu"},
		{"label implicit", "cpu", "label=cpu"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel := ParseSelector(tt.expr)
			got := sel.String()
			if got != tt.want {
				t.Errorf("ParseSelector(%q).String() = %q, want %q", tt.expr, got, tt.want)
			}
		})
	}
}

func TestRouter_Route(t *testing.T) {
	router := NewRouter()

	// Create two routes
	w1 := stream.NewFixedWindow(10)
	w2 := stream.NewFixedWindow(10)

	router.AddRoute(&Route{
		Selector: NewFieldSelector("label", "cpu"),
		Window:   w1,
	})

	router.AddRoute(&Route{
		Selector: NewFieldSelector("label", "memory"),
		Window:   w2,
	})

	// Route CPU point
	cpuPoint := stream.NewLabeledDataPoint("cpu", 45)
	matched := router.Route(cpuPoint)

	if matched != 1 {
		t.Errorf("expected 1 match for cpu, got %d", matched)
	}

	if w1.Len() != 1 {
		t.Errorf("w1 should have 1 point, got %d", w1.Len())
	}

	if w2.Len() != 0 {
		t.Errorf("w2 should have 0 points, got %d", w2.Len())
	}

	// Route memory point
	memPoint := stream.NewLabeledDataPoint("memory", 1024)
	matched = router.Route(memPoint)

	if matched != 1 {
		t.Errorf("expected 1 match for memory, got %d", matched)
	}

	if w2.Len() != 1 {
		t.Errorf("w2 should have 1 point, got %d", w2.Len())
	}
}
