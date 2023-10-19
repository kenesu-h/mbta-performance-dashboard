package pastalerts

import "database/sql"

// An ApiResponse represents a response from the MBTA Performance API's past alerts endpoint.
type ApiResponse struct {
  PastAlerts []PastAlert `json:"past_alerts"`
}

// A PastAlert represents an alert that was active during a time period in the past.
type PastAlert struct {
  ID            string         `json:"alert_id"`
  AlertVersions []AlertVersion `json:"alert_versions"`
}

// An AlertVersion represents a single version for an alert. A new one is created by the MBTA if any
// of its properties change.
type AlertVersion struct {
  AlertID          string           `json:"alert_id"`
  ID               string           `json:"version_id"`
  ValidFrom        string           `json:"valid_from"`
  ValidTo          string           `json:"valid_to"`
  Cause            string           `json:"cause"`
  Effect           string           `json:"effect"`
  HeaderText       string           `json:"header_text"`
  DescriptionText  string           `json:"description_text"`
  InformedEntities []InformedEntity `json:"informed_entity"`
  ActivePeriod     []ActivePeriod   `json:"active_period"`
}

// An InformedEntity represents services or facilities an alert version affects.
type InformedEntity struct {
  AlertID   string         `json:"alert_id"`
  VersionID string         `json:"version_id"`
  AgencyID  sql.NullString `json:"agency_id"`
  RouteID   sql.NullString `json:"route_id"`
  RouteType sql.NullString `json:"route_type"`
  TripID    sql.NullString `json:"trip_id"`
  StopID    sql.NullString `json:"stop_id"`
}

// An ActivePeriod represents the period for which an alert version is in effect.
type ActivePeriod struct {
  AlertID   string         `json:"alert_id"`
  VersionID string         `json:"version_id"`
  Start     string         `json:"start"`
  End       sql.NullString `json:"end"`
}
