package types

import (
	"database/sql"
  "fmt"
	"sync"
	"time"
)

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

type APIResponse[T Entity] interface {
  Entities() []T
}

// An EntityService represents a service that fetches entities from the MBTA Performance API and
// stores them within the local database.
type EntityService[T Entity] interface {
  BeginTx() (*sql.Tx, error)
  Lock()
  Unlock()

  FetchFromAPI(tx *sql.Tx, stopIDs []string, routeID string) ([]T, []error)
  Insert(tx *sql.Tx, entities []T) error
  Select(tx *sql.Tx, stopIDs []string, routeID string) ([]T, error)
  UpdateCacheDatetimes(tx *sql.Tx, stopIDs []string, routeID string) error
  DeleteOutdated(tx *sql.Tx) error
}

type BaseService struct {
  DB *sql.DB
  Mu *sync.Mutex
}

func (s *BaseService) BeginTx() (*sql.Tx, error) {
  tx, err := s.DB.Begin()
  if err != nil {
    return nil, fmt.Errorf("Error beginning transaction: %w", err)
  }
  return tx, nil
}

func (s *BaseService) Lock() {
  s.Mu.Lock()
}

func (s *BaseService) Unlock() {
  s.Mu.Unlock()
}

type Entity interface {
  StopID() string
  SetStopID(stopID string)
  RouteID() string
}

// A BaseEntity represents an entity corresponding to a stop and route ID combination.
type BaseEntity struct {
  StopID  string `json:"stop_id"`
  RouteID string `json:"route_id"`
}
