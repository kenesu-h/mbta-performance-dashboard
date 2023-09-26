package headways

// An ApiResponse represents a response from the MBTA Performance API's headways endpoint.
type ApiResponse struct {
	Headways []Headway `json:"headways"`
}

// A Headway represents the time between the previous and current trains' departures at a stop.
type Headway struct {
	StopID                  string `json:"stop_id"`
	RouteID                 string `json:"route_id"`
	PrevRouteID             string `json:"prev_route_id"`
	Direction               string `json:"direction"`
	CurrentDepDt            string `json:"current_dep_dt"`
	PreviousDepDt           string `json:"previous_dep_dt"`
	HeadwayTimeSec          string `json:"headway_time_sec"`
	BenchmarkHeadwayTimeSec string `json:"benchmark_headway_time_sec"`
}
