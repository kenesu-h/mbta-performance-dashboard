package dwells

// An ApiResponse represents a response from the MBTA Performance API's dwells endpoint.
type ApiResponse struct {
	Dwells []Dwell `json:"dwell_times"`
}

// A Dwell represents the time a train was stationary at a stop.
type Dwell struct {
	StopID       string `json:"stop_id"`
	RouteID      string `json:"route_id"`
	Direction    string `json:"direction"`
	ArrDt        string `json:"arr_dt"`
	DepDt        string `json:"dep_dt"`
	DwellTimeSec string `json:"dwell_time_sec"`
}
