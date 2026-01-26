package stream

import (
	"testing"
	"time"
)

func TestNewFixedWindow(t *testing.T) {
	w := NewFixedWindow(5)
	if w.Len() != 0 {
		t.Errorf("expected empty window, got %d points", w.Len())
	}
}

func TestWindow_AddAndLen(t *testing.T) {
	w := NewFixedWindow(10)

	for i := 0; i < 5; i++ {
		w.Add(NewDataPoint(float64(i)))
	}

	if w.Len() != 5 {
		t.Errorf("expected 5 points, got %d", w.Len())
	}
}

func TestWindow_FixedSizeEviction(t *testing.T) {
	w := NewFixedWindow(3)

	// Add 5 points to a window of size 3
	for i := 1; i <= 5; i++ {
		w.Add(NewDataPoint(float64(i)))
	}

	// Should only have 3 points (the last 3)
	if w.Len() != 3 {
		t.Errorf("expected 3 points after eviction, got %d", w.Len())
	}

	points := w.Points()
	expected := []float64{3, 4, 5}
	for i, p := range points {
		if p.Value != expected[i] {
			t.Errorf("point %d: expected %.0f, got %.0f", i, expected[i], p.Value)
		}
	}
}

func TestWindow_TimeBasedEviction(t *testing.T) {
	w := NewTimeWindow(100 * time.Millisecond)

	// Add a point
	w.Add(NewDataPoint(1))
	if w.Len() != 1 {
		t.Errorf("expected 1 point, got %d", w.Len())
	}

	// Wait for the time window to expire
	time.Sleep(150 * time.Millisecond)

	// Add another point, which should trigger eviction of the old one
	w.Add(NewDataPoint(2))

	if w.Len() != 1 {
		t.Errorf("expected 1 point after time eviction, got %d", w.Len())
	}

	points := w.Points()
	if len(points) > 0 && points[0].Value != 2 {
		t.Errorf("expected value 2, got %.0f", points[0].Value)
	}
}

func TestWindow_MinMax(t *testing.T) {
	tests := []struct {
		name    string
		values  []float64
		wantMin float64
		wantMax float64
	}{
		{
			name:    "single value",
			values:  []float64{42},
			wantMin: 42,
			wantMax: 42,
		},
		{
			name:    "ascending",
			values:  []float64{1, 2, 3, 4, 5},
			wantMin: 1,
			wantMax: 5,
		},
		{
			name:    "descending",
			values:  []float64{5, 4, 3, 2, 1},
			wantMin: 1,
			wantMax: 5,
		},
		{
			name:    "with negatives",
			values:  []float64{-10, 0, 10},
			wantMin: -10,
			wantMax: 10,
		},
		{
			name:    "all same",
			values:  []float64{7, 7, 7},
			wantMin: 7,
			wantMax: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewFixedWindow(100)
			for _, v := range tt.values {
				w.Add(NewDataPoint(v))
			}

			min, max := w.Scale()
			if min != tt.wantMin {
				t.Errorf("Min() = %v, want %v", min, tt.wantMin)
			}
			if max != tt.wantMax {
				t.Errorf("Max() = %v, want %v", max, tt.wantMax)
			}
		})
	}
}

func TestWindow_Last(t *testing.T) {
	w := NewFixedWindow(10)

	// Empty window
	_, ok := w.Last()
	if ok {
		t.Error("expected false for empty window")
	}

	// Add points
	w.Add(NewDataPoint(1))
	w.Add(NewDataPoint(2))
	w.Add(NewDataPoint(3))

	last, ok := w.Last()
	if !ok {
		t.Error("expected true after adding points")
	}
	if last.Value != 3 {
		t.Errorf("expected last value 3, got %.0f", last.Value)
	}
}

func TestWindow_Clear(t *testing.T) {
	w := NewFixedWindow(10)

	for i := 0; i < 5; i++ {
		w.Add(NewDataPoint(float64(i)))
	}

	w.Clear()

	if w.Len() != 0 {
		t.Errorf("expected 0 points after clear, got %d", w.Len())
	}

	min, max := w.Scale()
	if min != 0 || max != 0 {
		t.Errorf("expected min/max 0 after clear, got %.0f/%.0f", min, max)
	}
}

func TestWindow_Points_IsCopy(t *testing.T) {
	w := NewFixedWindow(10)
	w.Add(NewDataPoint(1))
	w.Add(NewDataPoint(2))

	points := w.Points()
	points[0].Value = 999 // modify the copy

	// Original should be unchanged
	original := w.Points()
	if original[0].Value != 1 {
		t.Error("Points() should return a copy, not the original slice")
	}
}
