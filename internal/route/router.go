package route

import (
	"github.com/danqzq/rift/internal/chart"
	"github.com/danqzq/rift/internal/stream"
)

// Route defines a routing rule from selector to chart.
type Route struct {
	Selector  Selector
	ChartType string
	Chart     chart.Chart
	Window    *stream.Window
}

// Router manages multiple routes and dispatches data points.
type Router struct {
	routes []*Route
}

// NewRouter creates a new router.
func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}

// AddRoute adds a routing rule.
func (r *Router) AddRoute(route *Route) {
	r.routes = append(r.routes, route)
}

// Route dispatches a data point to matching routes. Returns matched routes count.
func (r *Router) Route(p stream.DataPoint) int {
	matched := 0
	for _, route := range r.routes {
		if route.Selector.Matches(p) {
			route.Window.Add(p)
			matched++
		}
	}
	return matched
}

// Routes returns all configured routes.
func (r *Router) Routes() []*Route {
	return r.routes
}
