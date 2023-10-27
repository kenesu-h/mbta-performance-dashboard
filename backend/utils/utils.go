package utils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
  "strconv"
	"strings"
  "sync"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/mbta-performance-dashboard/consts"
	"github.com/mbta-performance-dashboard/types"
)

// Thanks to https://stackoverflow.com/a/50652930
// PgPlaceholders generates placeholders for a Postgres statement in the range (start, end].
//
// Postgres placeholders are represented as $i, where i is a one-based index.
func PgPlaceholders(start int, end int) string {
	placeholders := make([]string, end-start)
	for i := 0; i < end-start; i++ {
		placeholders[i] = fmt.Sprintf("$%d", start+i+1)
	}
	return strings.Join(placeholders, ", ")
}

func SliceToAnySlice[T any](slice []T) []any {
	var anySlice []any = []any{}
	for i := 0; i < len(slice); i++ {
		anySlice = append(anySlice, slice[i])
	}
  return anySlice
}

// StartOfToday returns the start of today (where the start is 00:00:00) in EST.
func StartOfToday() (time.Time, error) {
	newYork, err := time.LoadLocation("America/New_York")
	if err != nil {
    return time.Time{}, errors.New(fmt.Sprintf("Error loading New York timezone: %v", err))
	}

	now := time.Now().In(newYork)
	year, month, day := now.Date()
	startOfToday := time.Date(year, month, day, 0, 0, 0, 0, newYork)

  return startOfToday, nil
}

// PropagateToResponse makes a JSON response that propagates a provided error as is.
func PropagateToResponse(c *gin.Context, err error) {
  c.JSON(http.StatusInternalServerError, gin.H{
    "data": fmt.Sprintf("%v", err),
  })
}

// ValidateIDs validates stop and route IDs. Returns an error if at least one is invalid.
func ValidateIDs(tx *sql.Tx, stopIDs []string, routeID string) error {
	if len(stopIDs) == 0 {
		return errors.New("At least one stop ID required")
	}

	if routeID == "" {
		return errors.New("Route ID required")
	}

	rows, err := tx.Query(
    fmt.Sprintf("SELECT * FROM stop WHERE id IN (%s)", PgPlaceholders(0, len(stopIDs))),
    SliceToAnySlice[string](stopIDs)...
  )
	if err != nil {
    return fmt.Errorf("Error querying stops: %w", err)
	}

	var stops []types.Stop = []types.Stop{}
	for rows.Next() {
		var stop types.Stop
		err := rows.Scan(
			&stop.ID,
			&stop.RouteID,
			&stop.Name,
			&stop.Latitude,
			&stop.Longitude,
		)
		if err != nil {
      return fmt.Errorf("Error scanning stops: %w", err)
		}
		stops = append(stops, stop)
	}
	rows.Close()

	if len(stops) != len(stopIDs) {
		return errors.New("Not all stop IDs are valid")
	}


  rows, err = tx.Query("SELECT * from route WHERE id = $1", routeID)
	if err != nil {
    return fmt.Errorf("Error querying routes: %w", err)
	}

	if !rows.Next() {
    return fmt.Errorf("Invalid route ID %s", routeID)
	}
	rows.Close()

  return nil
}

func FetchFromAPI[T types.Entity, U types.APIResponse[T]](
  tx *sql.Tx,
  stopIDs []string,
  routeID string,
  lastCacheDatetimeTable string,
  endpoint string,
) ([]T, []error) {
  datetimes, err := getLastCacheDatetimes(
    tx,
    stopIDs,
    routeID,
    lastCacheDatetimeTable,
  )
  if err != nil {
    return nil, []error{ err }
  }

  startOfToday, err := StartOfToday()
  if err != nil {
    return nil, []error{ err }
  }
	endOfYesterday := startOfToday.Add(-1 * time.Second)

  var wg sync.WaitGroup
	results := make(chan []T)
  errs := []error{}

	for i := 0; i < len(stopIDs); i++ {
		wg.Add(1)

		go func(stopID string) {
			defer wg.Done()

			datetime, datetimeOk := datetimes[stopID]
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
        var entities []T
        entities, errs = FetchFromRequest[T, U](
          client,
          endpoint,
          errs,
          map[string]string{
            "stop": stopID,
            "route": routeID,
            "from_datetime": strconv.FormatInt(startOfRange.Unix(), 10),
            "to_datetime": strconv.FormatInt(endOfRange.Unix(), 10),
          },
        )
        for j := 0; j < len(entities); j++ {
          entities[j].SetStopID(stopID)
        }
				results <- entities

				startOfRange = startOfRange.AddDate(0, 0, 7)
				if endOfRange.Equal(endOfYesterday) {
					break
				} else if endOfRange.AddDate(0, 0, 7).Before(startOfToday) {
					endOfRange = endOfRange.AddDate(0, 0, 7)
				} else {
					endOfRange = endOfYesterday
				}
			}
		}(stopIDs[i])
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var entities []T = []T{}
	for result := range results {
		entities = append(entities, result...)
	}

  return entities, errs
}

func getLastCacheDatetimes(
  tx *sql.Tx,
  stopIDs []string,
  routeID string,
  table string,
) (map[string]time.Time, error) {
  rows, err := tx.Query(
    fmt.Sprintf(
      "SELECT stop_id, route_id, value AT TIME ZONE 'America/New_York' FROM %s WHERE stop_id IN " +
        "(%s) AND route_id = %s",
      table,
      PgPlaceholders(0, len(stopIDs)),
      PgPlaceholders(len(stopIDs), len(stopIDs)+1),
    ),
    SliceToAnySlice[string](append(stopIDs, routeID))...
  )
	if err != nil {
    return nil, fmt.Errorf("Error querying datetimes: %w", err)
	}

	var datetimes map[string]time.Time = make(map[string]time.Time)
	for rows.Next() {
		var datetime types.LastCacheDatetime
		err = rows.Scan(
			&datetime.StopID,
			&datetime.RouteID,
			&datetime.Value,
		)
		if err != nil {
      return nil, fmt.Errorf("Error scanning datetimes: %w", err)
		}
		datetimes[datetime.StopID] = datetime.Value
	}
	rows.Close()

  return datetimes, nil
}

func FetchFromRequest[T types.Entity, U types.APIResponse[T]](
  client http.Client,
  endpoint string,
  errs []error,
  params map[string]string,
) ([]T, []error) {
  req, err := http.NewRequest(
    "GET",
    fmt.Sprintf("%s/%s", consts.ApiPerformance, endpoint),
    nil,
  )
  if err != nil {
    errs = append(errs, fmt.Errorf("Error creating HTTP request to endpoint: %w", err))
    return nil, errs
  }

  query := req.URL.Query()
  query.Add("api_key", os.Getenv("PERFORMANCE_API_KEY"))
  query.Add("format", "json")
  for k, v := range params {
    query.Add(k, v)
  }
  req.URL.RawQuery = query.Encode()

  res, err := client.Do(req)
  if err != nil {
    errs = append(errs, fmt.Errorf("Error fetching entities: %w", err))
    return nil, errs
  }

  body, err := io.ReadAll(res.Body)
  if err != nil {
    errs = append(errs, fmt.Errorf("Errors reading response body: %w", err))
    return nil, errs
  }
  res.Body.Close()

  var apiRes U
  json.Unmarshal(body, &apiRes)
  return apiRes.Entities(), errs
}

func Cache[T types.Entity](c *gin.Context, service types.EntityService[T]) {
	stopIDs := strings.Split(c.DefaultQuery("stop_ids", ""), ",")
	routeID := c.DefaultQuery("route_id", "")

  err := func() error {
    tx, err := service.BeginTx()
    if err != nil {
      return err 
    }
    defer func() {
      if tx != nil {
        tx.Rollback()
      }
    }()

    service.Lock()
    defer service.Unlock()

    if err := ValidateIDs(tx, stopIDs, routeID); err != nil {
      return err
    }

    entities, errs := service.FetchFromAPI(tx, stopIDs, routeID)
    if len(errs) > 0 {
      return errors.Join(errs...)
    }

    if err = service.Insert(tx, entities); err != nil {
      return err
    }

    if err = service.UpdateCacheDatetimes(tx, stopIDs, routeID); err != nil {
      return err
    }

    if err = service.DeleteOutdated(tx); err != nil {
      return err
    }

    if err = tx.Commit(); err != nil {
      return fmt.Errorf("Error committing transaction: %w", err)
    }
    tx = nil

    return nil
  }()
  if err != nil {
    PropagateToResponse(c, err)
    return
  }

	c.JSON(http.StatusOK, gin.H{
		"data": "Successfully cached new entities",
	})
}

func Select[T types.Entity](c *gin.Context, service types.EntityService[T]) {
	stopIDs := strings.Split(c.DefaultQuery("stop_ids", ""), ",")
	routeID := c.DefaultQuery("route_id", "")

  var entities []T
  err := func() error {
    tx, err := service.BeginTx()
    if err != nil {
      return fmt.Errorf("Error beginning transaction: %w", err)
    }

    if err := ValidateIDs(tx, stopIDs, routeID); err != nil {
      return err
    }

    entities, err = service.Select(tx, stopIDs, routeID)
    if err != nil {
      return err
    }

    if err = tx.Commit(); err != nil {
      return fmt.Errorf("Error committing transaction: %w", err)
    }
    tx = nil;

    return nil
  }()
  if err != nil {
    PropagateToResponse(c, err)
    return
  }

	c.JSON(http.StatusOK, gin.H{
		"data": entities,
	})
}

func UpdateCacheDatetimes(
  tx *sql.Tx, 
  stopIDs []string, 
  routeID string, 
  lastCacheDatetimeTable string,
) error {
  datetimes, err := getLastCacheDatetimes(
    tx,
    stopIDs,
    routeID,
    lastCacheDatetimeTable,
  )
  if err != nil {
    return err
  }

  startOfToday, err := StartOfToday()
  if err != nil {
    return err
  }

  for i := 0; i < len(stopIDs); i++ {
    datetime, datetimeOk := datetimes[stopIDs[i]]
    if datetimeOk &&
    (datetime.After(startOfToday) ||
    datetime.Equal(startOfToday)) {
      continue
    }

    if !datetimeOk {
      _, err = tx.Exec(
        fmt.Sprintf(
          "INSERT INTO %s (stop_id, route_id, value) VALUES ($1, $2, $3)", 
          lastCacheDatetimeTable,
        ), 
        stopIDs[i], 
        routeID, 
        startOfToday,
      )
    } else {
      _, err = tx.Exec(
        fmt.Sprintf(
          "UPDATE %s SET value = $1 WHERE stop_id = $2 AND route_id = $3",
          lastCacheDatetimeTable,
        ),
        startOfToday, 
        stopIDs[i], 
        routeID,
      )
    }
    if err != nil {
      return fmt.Errorf("Error updating last cache datetime: %w", err)
    }
  }

  return nil
}

func DeleteOutdated(tx *sql.Tx, table string, dateColumn string) error {
  _, err := tx.Exec(fmt.Sprintf(
    "DELETE FROM %s WHERE %s < DATE_TRUNC('day', NOW() AT TIME ZONE 'America/New_York' - " +
      "INTERVAL '%d days')",
    table,
    dateColumn,
    consts.MaxDays,
  ))

  if err != nil {
    return fmt.Errorf("Error deleting outdated entities: %w", err)
  }

  return nil
}
