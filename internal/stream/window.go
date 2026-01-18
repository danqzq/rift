package stream

import (
	"sync"
	"time"
)

// WindowConfig holds configuration for window behavior.
type WindowConfig struct {
	// MaxSize limits the number of points in a fixed-size window (0 means no limit)
	MaxSize int

	// TimeWindow limits points to those within this duration (0 means no limit)
	TimeWindow time.Duration
}

// Window manages a sliding window of data points with auto-scaling (thread-safe)
type Window struct {
	mu     sync.RWMutex
	points []DataPoint
	config WindowConfig

	min float64
	max float64
}

// NewWindow creates a new Window with the given configuration.
func NewWindow(config WindowConfig) *Window {
	capacity := config.MaxSize
	if capacity <= 0 {
		capacity = 1000 // default capacity for time-based windows
	}
	return &Window{
		points: make([]DataPoint, 0, capacity),
		config: config,
	}
}

// NewFixedWindow creates a window that holds the last n points.
func NewFixedWindow(size int) *Window {
	return NewWindow(WindowConfig{MaxSize: size})
}

// NewTimeWindow creates a window that holds points from the last duration.
func NewTimeWindow(d time.Duration) *Window {
	return NewWindow(WindowConfig{TimeWindow: d})
}

// Add inserts a new data point into the window.
func (w *Window) Add(p DataPoint) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.points = append(w.points, p)
	w.evictLocked()
	w.updateScaleLocked()
}

// evictLocked removes points that fall outside the window bounds (must be called with mu held)
func (w *Window) evictLocked() {
	if w.config.MaxSize > 0 && len(w.points) > w.config.MaxSize {
		excess := len(w.points) - w.config.MaxSize
		w.points = w.points[excess:]
	}

	if w.config.TimeWindow > 0 {
		cutoff := time.Now().Add(-w.config.TimeWindow)
		idx := 0
		for idx < len(w.points) && w.points[idx].Timestamp.Before(cutoff) {
			idx++
		}
		if idx > 0 {
			w.points = w.points[idx:]
		}
	}
}

// updateScaleLocked recalculates min/max from current points (must be called with mu held)
func (w *Window) updateScaleLocked() {
	if len(w.points) == 0 {
		w.min = 0
		w.max = 0
		return
	}

	w.min = w.points[0].Value
	w.max = w.points[0].Value

	for _, p := range w.points[1:] {
		if p.Value < w.min {
			w.min = p.Value
		}
		if p.Value > w.max {
			w.max = p.Value
		}
	}
}

// Points returns a copy of all points in the window
func (w *Window) Points() []DataPoint {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make([]DataPoint, len(w.points))
	copy(result, w.points)
	return result
}

// Len returns the number of points currently in the window
func (w *Window) Len() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.points)
}

// Min returns the minimum value in the current window.
func (w *Window) Min() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.min
}

// Max returns the maximum value in the current window.
func (w *Window) Max() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.max
}

// Scale returns both min and max for efficient access.
func (w *Window) Scale() (min, max float64) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.min, w.max
}

// Clear removes all points from the window.
func (w *Window) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.points = w.points[:0]
	w.min = 0
	w.max = 0
}

// Last returns the most recent data point, or an empty DataPoint if the window is empty.
func (w *Window) Last() (DataPoint, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if len(w.points) == 0 {
		return DataPoint{}, false
	}
	return w.points[len(w.points)-1], true
}
