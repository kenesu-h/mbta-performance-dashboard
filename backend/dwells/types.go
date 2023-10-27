package dwells

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"

	"github.com/lib/pq"
	"github.com/mbta-performance-dashboard/types"
	"github.com/mbta-performance-dashboard/utils"
)

// An APIResponse represents a response from the MBTA Performance API's dwells endpoint.
type APIResponse struct {
	Dwells []*Dwell `json:"dwell_times"`
}

func (a *APIResponse) Entities() []*Dwell {
  return a.Dwells
}


// A Dwell represents the time a train was stationary at a stop.
type Dwell struct {
	types.BaseEntity
	Direction    string `json:"direction"`
	ArrDt        string `json:"arr_dt"`
	DepDt        string `json:"dep_dt"`
	DwellTimeSec string `json:"dwell_time_sec"`
}

func (d *Dwell) StopID() string {
  return d.BaseEntity.StopID
}

func (d *Dwell) SetStopID(stopID string) {
  d.BaseEntity.StopID = stopID
}

func (d *Dwell) RouteID() string {
  return d.BaseEntity.RouteID
}


type DwellService struct {
  types.BaseService
}

func NewService(db *sql.DB, mu *sync.Mutex) *DwellService {
  return &DwellService { BaseService: types.BaseService{ DB: db, Mu: mu } }
}

func (s *DwellService) BeginTx() (*sql.Tx, error) {
  return s.BaseService.BeginTx()
}

func (s *DwellService) Lock() {
  s.BaseService.Lock()
}

func (s *DwellService) Unlock() {
  s.BaseService.Unlock()
}

func (s *DwellService) FetchFromAPI(
  tx *sql.Tx,
  stopIDs []string,
  routeID string,
) ([]*Dwell, []error) {
  return utils.FetchFromAPI[*Dwell, *APIResponse](
    tx,
    stopIDs,
    routeID,
    "last_dwell_cache_datetime",
    "dwells",
  )
}

func (s *DwellService) Insert(tx *sql.Tx, dwells []*Dwell) error {
	if len(dwells) == 0 {
    return nil
	}

  var paramStopIDs []string = []string{}
  var paramRouteIDs []string = []string{}
  var paramDirections []bool = []bool{}
  var paramArrDts []int = []int{}
  var paramDepDts []int = []int{}
  var paramDwellTimeSecs []int = []int{}

  for i := 0; i < len(dwells); i++ {
		convertedDirection, err := strconv.Atoi(dwells[i].Direction)
		if err != nil {
			panic(fmt.Sprintf("Error converting direction to integer: %v", err))
		}
		convertedArrDt, err := strconv.Atoi(dwells[i].ArrDt)
		if err != nil {
			panic(fmt.Sprintf("Error converting arrival datetime to integer: %v", err))
		}
		convertedDepDt, err := strconv.Atoi(dwells[i].DepDt)
		if err != nil {
			panic(fmt.Sprintf("Error converting departure datetime to integer: %v", err))
		}
		convertedDwellTimeSec, err := strconv.Atoi(dwells[i].DwellTimeSec)
		if err != nil {
			panic(fmt.Sprintf("Error converting dwell time to integer: %v", err))
		}

    paramStopIDs = append(paramStopIDs, dwells[i].StopID())
    paramRouteIDs = append(paramRouteIDs, dwells[i].RouteID())
    paramDirections = append(paramDirections, convertedDirection == 1)
    paramArrDts = append(paramArrDts, convertedArrDt)
    paramDepDts = append(paramDepDts, convertedDepDt)
    paramDwellTimeSecs = append(paramDwellTimeSecs, convertedDwellTimeSec)
	}

  _, err := tx.Exec(
    "INSERT INTO dwell (stop_id, route_id, direction, arr_dt, dep_dt, dwell_time_sec) " +
      "SELECT " +
      "unnest($1::text[]) AS stop_id, " +
      "unnest($2::text[]) AS route_id, " +
      "unnest($3::boolean[]) AS direction, " +
      "TO_TIMESTAMP(unnest($4::int[])) AS arr_dt, " +
      "TO_TIMESTAMP(unnest($5::int[])) AS dep_dt, " +
      "unnest($6::int[]) AS dwell_time_sec",
    pq.Array(paramStopIDs),
    pq.Array(paramRouteIDs),
    pq.Array(paramDirections),
    pq.Array(paramArrDts),
    pq.Array(paramDepDts),
    pq.Array(paramDwellTimeSecs),
  )
	if err != nil {
    return fmt.Errorf("Error inserting dwells: %w", err)
	}

  return nil
}

func (s *DwellService) Select(
  tx *sql.Tx,
  stopIDs []string,
  routeID string,
) ([]*Dwell, error) {
	rows, err := tx.Query(
    fmt.Sprintf(
      "SELECT stop_id, route_id, direction, arr_dt AT TIME ZONE 'America/New_York', dep_dt AT " +
        "TIME ZONE 'America/New_York', dwell_time_sec FROM dwell WHERE stop_id IN (%s) AND " +
        "route_id = %s",
      utils.PgPlaceholders(0, len(stopIDs)),
      utils.PgPlaceholders(len(stopIDs), len(stopIDs)+1),
    ),
    utils.SliceToAnySlice[string](append(stopIDs, routeID))...,
  )
	if err != nil {
    return nil, fmt.Errorf("Error fetching dwells: %w", err)
	}

	var dwells []*Dwell = []*Dwell{}
	for rows.Next() {
		var dwell Dwell
		err := rows.Scan(
			&dwell.BaseEntity.StopID,
			&dwell.BaseEntity.RouteID,
			&dwell.Direction,
			&dwell.ArrDt,
			&dwell.DepDt,
			&dwell.DwellTimeSec,
		)
		if err != nil {
      return nil, fmt.Errorf("Error scanning dwells: %w", err)
		}
		dwells = append(dwells, &dwell)
	}
	rows.Close()

	return dwells, nil
}

func (s *DwellService) UpdateCacheDatetimes(tx *sql.Tx, stopIDs []string, routeID string) error {
  return utils.UpdateCacheDatetimes(tx, stopIDs, routeID, "last_dwell_cache_datetime")
}

func (s *DwellService) DeleteOutdated(tx *sql.Tx) error {
  return utils.DeleteOutdated(tx, "dwell", "arr_dt")
}
