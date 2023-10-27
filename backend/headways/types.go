package headways

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"

	"github.com/lib/pq"
	"github.com/mbta-performance-dashboard/types"
	"github.com/mbta-performance-dashboard/utils"
)

// An APIResponse represents a response from the MBTA Performance API's headways endpoint.
type APIResponse struct {
	Headways []*Headway `json:"headways"`
}

func (a *APIResponse) Entities() []*Headway {
  return a.Headways
}


// A Headway represents the time between the previous and current trains' departures at a stop.
type Headway struct {
	types.BaseEntity
	PrevRouteID             string `json:"prev_route_id"`
	Direction               string `json:"direction"`
	CurrentDepDt            string `json:"current_dep_dt"`
	PreviousDepDt           string `json:"previous_dep_dt"`
	HeadwayTimeSec          string `json:"headway_time_sec"`
	BenchmarkHeadwayTimeSec string `json:"benchmark_headway_time_sec"`
}

func (h *Headway) StopID() string {
  return h.BaseEntity.StopID
}

func (h *Headway) SetStopID(stopID string) {
  h.BaseEntity.StopID = stopID
}

func (h *Headway) RouteID() string {
  return h.BaseEntity.RouteID
}


type HeadwayService struct {
  types.BaseService
}

func NewService(db *sql.DB, mu *sync.Mutex) *HeadwayService {
  return &HeadwayService{ BaseService: types.BaseService{ DB: db, Mu: mu } }
}

func (s *HeadwayService) BeginTx() (*sql.Tx, error) {
  return s.BaseService.BeginTx()
}

func (s *HeadwayService) Lock() {
  s.BaseService.Lock()
}

func (s *HeadwayService) Unlock() {
  s.BaseService.Unlock()
}

func (s *HeadwayService) FetchFromAPI(
  tx *sql.Tx,
  stopIDs []string,
  routeID string,
) ([]*Headway, []error) {
  return utils.FetchFromAPI[*Headway, *APIResponse](
    tx,
    stopIDs,
    routeID,
    "last_headway_cache_datetime",
    "headways",
  )
}

func (s *HeadwayService) Insert(tx *sql.Tx, headways []*Headway) error {
	if len(headways) == 0 {
    return nil
	}

  var paramStopIDs []string = []string{}
  var paramRouteIDs []string = []string{}
  var paramPrevRouteIDs []string = []string{}
  var paramDirections []bool = []bool{}
  var paramCurrentDepDts []int = []int{}
  var paramPreviousDepDts []int = []int{}
  var paramHeadwayTimeSecs []int = []int{}
  var paramBenchmarkHeadwayTimeSecs []int = []int{}

  for i := 0; i < len(headways); i++ {
		convertedDirection, err := strconv.Atoi(headways[i].Direction)
		if err != nil {
      return fmt.Errorf("Error converting direction to integer: %w", err)
		}
		convertedCurrentDepDt, err := strconv.Atoi(headways[i].CurrentDepDt)
		if err != nil {
      return fmt.Errorf("Error converting current departure datetime to integer: %w", err)
		}
		convertedPreviousDepDt, err := strconv.Atoi(headways[i].PreviousDepDt)
		if err != nil {
      return fmt.Errorf("Error converting previous departure datetime to integer: %w", err)
		}
		convertedHeadwayTimeSec, err := strconv.Atoi(headways[i].HeadwayTimeSec)
		if err != nil {
      return fmt.Errorf("Error converting headway time to integer: %w", err)
		}
		convertedBenchmarkHeadwayTimeSec, err := strconv.Atoi(headways[i].BenchmarkHeadwayTimeSec)
		if err != nil {
      return fmt.Errorf("Error converting benchmark headway time to integer: %w", err)
		}

    paramStopIDs = append(paramStopIDs, headways[i].StopID())
    paramRouteIDs = append(paramRouteIDs, headways[i].RouteID())
    paramPrevRouteIDs = append(paramPrevRouteIDs, headways[i].PrevRouteID)
    paramDirections = append(paramDirections, convertedDirection == 1)
    paramCurrentDepDts = append(paramCurrentDepDts, convertedCurrentDepDt)
    paramPreviousDepDts = append(paramPreviousDepDts, convertedPreviousDepDt)
    paramHeadwayTimeSecs = append(paramHeadwayTimeSecs, convertedHeadwayTimeSec)
    paramBenchmarkHeadwayTimeSecs = append(
      paramBenchmarkHeadwayTimeSecs,
      convertedBenchmarkHeadwayTimeSec,
    )
	}

  _, err := tx.Exec(
    "INSERT INTO headway (stop_id, route_id, prev_route_id, direction, " +
      "current_dep_dt, previous_dep_dt, headway_time_sec, benchmark_headway_time_sec) SELECT " +
      "unnest($1::text[]) AS stop_id, " +
      "unnest($2::text[]) AS route_id, " +
      "unnest($3::text[]) AS prev_route_id, " +
      "unnest($4::boolean[]) AS direction, " +
      "TO_TIMESTAMP(unnest($5::int[])) AS current_dep_dt, " +
      "TO_TIMESTAMP(unnest($6::int[])) AS previous_dep_dt, " +
      "unnest($7::int[]) AS headway_time_sec, " +
      "unnest($8::int[]) AS benchmark_headway_time_sec",
    pq.Array(paramStopIDs),
    pq.Array(paramRouteIDs),
    pq.Array(paramPrevRouteIDs),
    pq.Array(paramDirections),
    pq.Array(paramCurrentDepDts),
    pq.Array(paramPreviousDepDts),
    pq.Array(paramHeadwayTimeSecs),
    pq.Array(paramBenchmarkHeadwayTimeSecs),
  )
	if err != nil {
    return fmt.Errorf("Error inserting headways: %w", err)
	}

  return nil
}

func (s *HeadwayService) Select(
  tx *sql.Tx,
  stopIDs []string,
  routeID string,
) ([]*Headway, error) {
	rows, err := tx.Query(
    fmt.Sprintf(
      "SELECT stop_id, route_id, prev_route_id, direction, current_dep_dt AT TIME " +
        "ZONE 'America/New_York', previous_dep_dt AT TIME ZONE 'America/New_York', " +
        "headway_time_sec, benchmark_headway_time_sec FROM headway WHERE stop_id IN (%s) AND " +
        "route_id = %s",
      utils.PgPlaceholders(0, len(stopIDs)),
      utils.PgPlaceholders(len(stopIDs), len(stopIDs)+1),
    ),
    utils.SliceToAnySlice[string](append(stopIDs, routeID))...,
  )
	if err != nil {
    return nil, fmt.Errorf("Error fetching headways: %w", err)
	}

	var headways []*Headway = []*Headway{}
	for rows.Next() {
		var headway Headway
		err := rows.Scan(
			&headway.BaseEntity.StopID,
			&headway.BaseEntity.RouteID,
			&headway.PrevRouteID,
			&headway.Direction,
			&headway.CurrentDepDt,
			&headway.PreviousDepDt,
			&headway.HeadwayTimeSec,
			&headway.BenchmarkHeadwayTimeSec,
		)
		if err != nil {
      return nil, fmt.Errorf("Error scanning headways: %w", err)
		}
		headways = append(headways, &headway)
	}
	rows.Close()

	return headways, nil
}

func (s *HeadwayService) UpdateCacheDatetimes(tx *sql.Tx, stopIDs []string, routeID string) error {
  return utils.UpdateCacheDatetimes(tx, stopIDs, routeID, "last_headway_cache_datetime")
}

func (s *HeadwayService) DeleteOutdated(tx *sql.Tx) error {
  return utils.DeleteOutdated(tx, "headway", "current_dep_dt")
}
