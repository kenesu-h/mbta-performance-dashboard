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

// An APIResponse represents a response from the MBTA Performance API that contains a list of
// generic entities.
type APIResponse[T Entity] interface {
  Entities() []T
}

// An EntityService represents a service that fetches generic entities from the MBTA Performance API
// and interacts with a database to store them.
type EntityService[T Entity] interface {
  // BeginTx begins a database transaction.
  BeginTx() (*sql.Tx, error)

  // Lock locks the service's mutex.
  //
  // Best used to ensure that there are no data races when caching new entities.
  Lock()

  // Unlock unlocks the service's mutex.
  Unlock()

  // FetchFromAPI fetches this service's entities from the MBTA Performance API.
  FetchFromAPI(tx *sql.Tx, stopIDs []string, routeID string) ([]T, []error)

  // Insert inserts provided entities into the database.
  Insert(tx *sql.Tx, entities []T) error

  // Select selects entities from the database whose stop ID matches one of the provided stop IDs,
  // and whose route ID matches as well.
  Select(tx *sql.Tx, stopIDs []string, routeID string) ([]T, error)

  // UpdateCacheDatetimes updates this service's entities' last cache datetimes to the start of
  // today.
  UpdateCacheDatetimes(tx *sql.Tx, stopIDs []string, routeID string) error

  // DeleteOutdated deletes this service's entities whose dates are set to before 30 days ago.
  DeleteOutdated(tx *sql.Tx) error
}

// A BaseService represents a basic EntityService, which must have a way to interact with the
// database and a mutex to prevent data races.
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

// An Entity represents a generic entity from the MBTA Performance API.
type Entity interface {
  StopID() string
  SetStopID(stopID string)
  RouteID() string
}

// A BaseEntity represents an generic entity corresponding to a stop and route ID combination.
//
// Not all entities will have just a single stop ID, but all entities will have a route ID.
type BaseEntity struct {
  StopID  string `json:"stop_id"`
  RouteID string `json:"route_id"`
}
