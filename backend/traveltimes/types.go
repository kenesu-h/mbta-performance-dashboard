package traveltimes

import (
	"database/sql"
  "errors"
	"fmt"
  "net/http"
	"strconv"
	"sync"
  "time"

	"github.com/lib/pq"
	"github.com/mbta-performance-dashboard/consts"
	"github.com/mbta-performance-dashboard/types"
	"github.com/mbta-performance-dashboard/utils"
)

// An APIResponse represents a response from the MBTA Performance API's travel times endpoint.
type APIResponse struct {
	TravelTimes []*TravelTime `json:"travel_times"`
}

func (a *APIResponse) Entities() []*TravelTime {
  return a.TravelTimes
}


// A TravelTime represents the travel time of a train from an origin to a destination.
type TravelTime struct {
  types.BaseEntity
	FromStopID             string `json:"from_stop_id"`
	ToStopID               string `json:"to_stop_id"`
	Direction              string `json:"direction"`
	DepDt                  string `json:"dep_dt"`
	ArrDt                  string `json:"arr_dt"`
	TravelTimeSec          string `json:"travel_time_sec"`
	BenchmarkTravelTimeSec string `json:"benchmark_travel_time_sec"`
}

// These don't actually matter
func (t *TravelTime) StopID() string {
  return t.BaseEntity.StopID
}

func (t *TravelTime) SetStopID(stopID string) {
  t.BaseEntity.StopID = stopID
}

// This does though
func (t *TravelTime) RouteID() string {
  return t.BaseEntity.RouteID
}


// A LastCacheDatetime represents the last time data was cached for this origin-destination-route ID
// combination.
//
// A different LastCacheDatetime is needed to accommodate the fact that there are two stop IDs
// involved in the combination - not just one.
type LastCacheDatetime struct {
	FromStopID string    `json:"from_stop_id"`
	ToStopID   string    `json:"to_stop_id"`
	RouteID    string    `json:"route_id"`
	Value      time.Time `json:"value"`
}


// A TravelTimeService represents a service that will fetch and store travel times.
type TravelTimeService struct {
  types.BaseService
}

func NewService(db *sql.DB, mu *sync.Mutex) *TravelTimeService {
  return &TravelTimeService { BaseService: types.BaseService{ DB: db, Mu: mu } }
}

func (s *TravelTimeService) BeginTx() (*sql.Tx, error) {
  return s.BaseService.BeginTx()
}

func (s *TravelTimeService) Lock() {
  s.BaseService.Lock()
}

func (s *TravelTimeService) Unlock() {
  s.BaseService.Unlock()
}

func (s *TravelTimeService) FetchFromAPI(
  tx *sql.Tx,
  stopIDs []string,
  routeID string,
) ([]*TravelTime, []error) {
  return nil, []error{ errors.New("Please use FetchTravelTimesFromAPI instead") }
}

func (s *TravelTimeService) FetchTravelTimesFromAPI(
  tx *sql.Tx,
  fromStopIDs []string,
  toStopIDs []string,
  routeID string,
) ([]*TravelTime, []error) {
  datetimes, err := s.lastCacheDatetimes(
    tx,
    fromStopIDs,
    toStopIDs,
    routeID,
  )
  if err != nil {
    return nil, []error{ err }
  }

  startOfToday, err := utils.StartOfToday()
  if err != nil {
    return nil, []error{ err }
  }
	endOfYesterday := startOfToday.Add(-1 * time.Second)

  var wg sync.WaitGroup
	results := make(chan []*TravelTime)
  errs := []error{}

	for i := 0; i < len(fromStopIDs); i++ {
    for j := 0; j < len(toStopIDs); j++ {
      wg.Add(1)

      go func(fromStopID string, toStopID string) {
        defer wg.Done()

        datetime, datetimeOk := datetimes[fromStopID][toStopID]
        if datetimeOk &&
          (datetime.After(startOfToday) ||
            datetime.Equal(startOfToday)) {
          return
        }

        var startOfRange time.Time
        var endOfRange time.Time

        if !datetimeOk ||
          startOfToday.Sub(datetime).Hours()/24 >= float64(consts.MaxDays) {
          startOfRange = startOfToday.AddDate(0, 0, -consts.MaxDays)
        } else {
          startOfRange = datetime
        }

        if startOfRange.AddDate(0, 0, 7).Before(startOfToday) {
          endOfRange = startOfRange.AddDate(0, 0, 7).Add(-1 * time.Second)
        } else {
          endOfRange = endOfYesterday
        }

        client := http.Client{}

        for true {
          var travelTimes []*TravelTime
          travelTimes, errs = utils.FetchFromRequest[*TravelTime, *APIResponse](
            client,
            "traveltimes",
            errs,
            map[string]string{
              "from_stop": fromStopID,
              "to_stop": toStopID,
              "route": routeID,
              "from_datetime": strconv.FormatInt(startOfRange.Unix(), 10),
              "to_datetime": strconv.FormatInt(endOfRange.Unix(), 10),
            },
          )
          for k := 0; k < len(travelTimes); k++ {
            travelTimes[k].FromStopID = fromStopID
            travelTimes[k].ToStopID = toStopID
          }
          results <- travelTimes

          startOfRange = startOfRange.AddDate(0, 0, 7)
          if endOfRange.Equal(endOfYesterday) {
            break
          } else if endOfRange.AddDate(0, 0, 7).Before(startOfToday) {
            endOfRange = endOfRange.AddDate(0, 0, 7)
          } else {
            endOfRange = endOfYesterday
          }
        }
      }(fromStopIDs[i], toStopIDs[j])
    }
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var travelTimes []*TravelTime = []*TravelTime{}
	for result := range results {
		travelTimes = append(travelTimes, result...)
	}

  return travelTimes, errs
}

func (s *TravelTimeService) lastCacheDatetimes(
  tx *sql.Tx,
  fromStopIDs []string,
  toStopIDs []string,
  routeID string,
) (map[string]map[string]time.Time, error) {
  rows, err := tx.Query(
    fmt.Sprintf(
      "SELECT from_stop_id, to_stop_id, route_id, value AT TIME ZONE 'America/New_York' FROM " +
        "last_travel_time_cache_datetime WHERE from_stop_id IN (%s) AND to_stop_id IN (%s) AND " +
        "route_id = %s",
      utils.PgPlaceholders(0, len(fromStopIDs)),
      utils.PgPlaceholders(len(fromStopIDs), len(fromStopIDs)+len(toStopIDs)),
      utils.PgPlaceholders(len(fromStopIDs)+len(toStopIDs), len(fromStopIDs)+len(toStopIDs) + 1),
    ),
    utils.SliceToAnySlice[string](append(fromStopIDs, append(toStopIDs, routeID)...))...
  )
	if err != nil {
    return nil, fmt.Errorf("Error querying datetimes: %w", err)
	}

	var datetimes map[string]map[string]time.Time = make(map[string]map[string]time.Time)
	for rows.Next() {
		var datetime LastCacheDatetime
		err = rows.Scan(
			&datetime.FromStopID,
      &datetime.ToStopID,
			&datetime.RouteID,
			&datetime.Value,
		)
		if err != nil {
      return nil, fmt.Errorf("Error scanning datetimes: %w", err)
		}
    if datetimes[datetime.FromStopID] == nil {
      datetimes[datetime.FromStopID] = make(map[string]time.Time)
    }
    datetimes[datetime.FromStopID][datetime.ToStopID] = datetime.Value
	}
	rows.Close()

  return datetimes, nil
}

func (s *TravelTimeService) Insert(tx *sql.Tx, travelTimes []*TravelTime) error {
	if len(travelTimes) == 0 {
    return nil
	}

  var paramFromStopIDs []string = []string{}
  var paramToStopIDs []string = []string{}
  var paramRouteIDs []string = []string{}
  var paramDirections []bool = []bool{}
  var paramDepDts []int = []int{}
  var paramArrDts []int = []int{}
  var paramTravelTimeSecs []int = []int{}
  var paramBenchmarkTravelTimeSecs []int = []int{}

  for i := 0; i < len(travelTimes); i++ {
		convertedDirection, err := strconv.Atoi(travelTimes[i].Direction)
		if err != nil {
			panic(fmt.Sprintf("Error converting direction to integer: %v", err))
		}
		convertedDepDt, err := strconv.Atoi(travelTimes[i].DepDt)
		if err != nil {
			panic(fmt.Sprintf("Error converting departure datetime to integer: %v", err))
		}
		convertedArrDt, err := strconv.Atoi(travelTimes[i].ArrDt)
		if err != nil {
			panic(fmt.Sprintf("Error converting arrival datetime to integer: %v", err))
		}
		convertedTravelTimeSec, err := strconv.Atoi(travelTimes[i].TravelTimeSec)
		if err != nil {
			panic(fmt.Sprintf("Error converting travel time to integer: %v", err))
		}
		convertedBenchmarkTravelTimeSec, err := strconv.Atoi(travelTimes[i].BenchmarkTravelTimeSec)
		if err != nil {
			panic(fmt.Sprintf("Error converting benchmark travel time to integer: %v", err))
		}

    paramFromStopIDs = append(paramFromStopIDs, travelTimes[i].FromStopID)
    paramToStopIDs = append(paramToStopIDs, travelTimes[i].ToStopID)
    paramRouteIDs = append(paramRouteIDs, travelTimes[i].RouteID())
    paramDirections = append(paramDirections, convertedDirection == 1)
    paramDepDts = append(paramDepDts, convertedDepDt)
    paramArrDts = append(paramArrDts, convertedArrDt)
    paramTravelTimeSecs = append(paramTravelTimeSecs, convertedTravelTimeSec)
    paramBenchmarkTravelTimeSecs = append(
      paramBenchmarkTravelTimeSecs,
      convertedBenchmarkTravelTimeSec,
    )
	}

  _, err := tx.Exec(
    "INSERT INTO travel_time (from_stop_id, to_stop_id, route_id, direction, dep_dt, "+
      "arr_dt, travel_time_sec, benchmark_travel_time_sec) SELECT " +
      "unnest($1::text[]) AS from_stop_id, " +
      "unnest($2::text[]) AS to_stop_id, " +
      "unnest($3::text[]) AS route_id, " +
      "unnest($4::boolean[]) AS direction, " +
      "TO_TIMESTAMP(unnest($5::int[])) AS dep_dt, " +
      "TO_TIMESTAMP(unnest($6::int[])) AS arr_dt, " +
      "unnest($7::int[]) AS travel_time_sec, " +
      "unnest($8::int[]) AS benchmark_travel_time_sec",
    pq.Array(paramFromStopIDs),
    pq.Array(paramToStopIDs),
    pq.Array(paramRouteIDs),
    pq.Array(paramDirections),
    pq.Array(paramDepDts),
    pq.Array(paramArrDts),
    pq.Array(paramTravelTimeSecs),
    pq.Array(paramBenchmarkTravelTimeSecs),
  )
	if err != nil {
    return fmt.Errorf("Error inserting travel times: %w", err)
	}

  return nil
}

func (s *TravelTimeService) Select(
  tx *sql.Tx,
  stopIDs []string,
  routeID string,
) ([]*TravelTime, error) {
  return nil, errors.New("Please use SelectTravelTimes instead")
}

func (s *TravelTimeService) SelectTravelTimes(
  tx *sql.Tx,
  fromStopIDs []string,
  toStopIDs []string,
  routeID string,
) ([]*TravelTime, error) {
 	rows, err := tx.Query(
    fmt.Sprintf(
      "SELECT from_stop_id, to_stop_id, route_id, direction, dep_dt AT TIME ZONE " +
        "'America/New_York', arr_dt AT TIME ZONE 'America/New_York', travel_time_sec, " +
        "benchmark_travel_time_sec FROM travel_time WHERE from_stop_id IN (%s) AND to_stop_id IN " +
        "(%s) AND route_id = %s",
      utils.PgPlaceholders(0, len(fromStopIDs)),
      utils.PgPlaceholders(len(fromStopIDs), len(fromStopIDs)+len(toStopIDs)),
      utils.PgPlaceholders(len(fromStopIDs)+len(toStopIDs), len(fromStopIDs)+len(toStopIDs)+1),
    ),
    utils.SliceToAnySlice[string](append(fromStopIDs, append(toStopIDs, routeID)...))...,
  )
	if err != nil {
    return nil, fmt.Errorf("Error fetching travel times: %w", err)
	}

	var travelTimes []*TravelTime = []*TravelTime{}
	for rows.Next() {
		var travelTime TravelTime
		err := rows.Scan(
			&travelTime.FromStopID,
      &travelTime.ToStopID,
			&travelTime.BaseEntity.RouteID,
			&travelTime.Direction,
			&travelTime.DepDt,
			&travelTime.ArrDt,
			&travelTime.TravelTimeSec,
      &travelTime.BenchmarkTravelTimeSec,
		)
		if err != nil {
      return nil, fmt.Errorf("Error scanning travel times: %w", err)
		}
		travelTimes = append(travelTimes, &travelTime)
	}
	rows.Close()

	return travelTimes, nil 
}

func (s *TravelTimeService) UpdateCacheDatetimes(tx *sql.Tx, stopIDs []string, routeID string) error {
  return errors.New("Please use UpdateTravelTimeCacheDatetimes instead")
}

func (s *TravelTimeService) UpdateTravelTimeCacheDatetimes(
  tx *sql.Tx, 
  fromStopIDs []string,
  toStopIDs []string,
  routeID string,
) error {
  datetimes, err := s.lastCacheDatetimes(
    tx,
    fromStopIDs,
    toStopIDs,
    routeID,
  )
  if err != nil {
    return err
  }

  startOfToday, err := utils.StartOfToday()
  if err != nil {
    return err
  }

  for i := 0; i < len(fromStopIDs); i++ {
    for j := 0; j < len(toStopIDs); j++ {
      datetime, datetimeOk := datetimes[fromStopIDs[i]][toStopIDs[j]]
      if datetimeOk &&
      (datetime.After(startOfToday) ||
      datetime.Equal(startOfToday)) {
        continue
      }

      if !datetimeOk {
        _, err = tx.Exec(
          "INSERT INTO last_travel_time_cache_datetime (from_stop_id, to_stop_id, route_id, " +
            "value) VALUES ($1, $2, $3, $4)", 
          fromStopIDs[i],
          toStopIDs[j],
          routeID, 
          startOfToday,
        )
      } else {
        _, err = tx.Exec(
          "UPDATE last_travel_time_cache_datetime SET value = $1 WHERE from_stop_id = $2 AND " +
            "to_stop_id = $3 AND route_id = $4",
          startOfToday, 
          fromStopIDs[i], 
          toStopIDs[j],
          routeID,
        )
      }
      if err != nil {
        return fmt.Errorf("Error updating last cache datetime: %w", err)
      }
    }
  }

  return nil
}

func (s *TravelTimeService) DeleteOutdated(tx *sql.Tx) error {
  return utils.DeleteOutdated(tx, "travel_time", "dep_dt")
}
