package traveltimes

import "time"

// A LastCacheDatetime represents the last time data was cached for this origin-destination-route ID
// combination.
type LastCacheDatetime struct {
	FromStopID string    `json:"from_stop_id"`
	ToStopID   string    `json:"to_stop_id"`
	RouteID    string    `json:"route_id"`
	Value      time.Time `json:"value"`
}

// An ApiResponse represents a response from the MBTA Performance API's travel times endpoint.
type ApiResponse struct {
	TravelTimes []TravelTime `json:"travel_times"`
}

// A TravelTime represents the travel time of a train from an origin to a destination.
type TravelTime struct {
	FromStopID             string `json:"from_stop_id"`
	ToStopID               string `json:"to_stop_id"`
	RouteID                string `json:"route_id"`
	Direction              string `json:"direction"`
	DepDt                  string `json:"dep_dt"`
	ArrDt                  string `json:"arr_dt"`
	TravelTimeSec          string `json:"travel_time_sec"`
	BenchmarkTravelTimeSec string `json:"benchmark_travel_time_sec"`
}
