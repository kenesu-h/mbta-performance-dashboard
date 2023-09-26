package types

import "time"

// A Route represents a route on the MBTA, like train lines and buses.
type Route struct {
	ID string `json:"id"`
}

// A Shape represents a path for a route, represented by its polyline.
type Shape struct {
	ID       string `json:"id"`
	RouteID  string `json:"route_id"`
	Polyline string `json:"polyline"`
}

// A Stop represents a stop on a route.
type Stop struct {
	ID        string  `json:"id"`
	RouteID   string  `json:"route_id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// A LastCacheDatetime represents the last time data was cached for this stop ID-route ID
// combination.
type LastCacheDatetime struct {
	StopID  string    `json:"stop_id"`
	RouteID string    `json:"route_id"`
	Value   time.Time `json:"value"`
}
