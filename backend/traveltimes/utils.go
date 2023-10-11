package traveltimes

import (
	"fmt"
	"io"
  "log"
	"os"
	"strconv"
	"strings"
  "sync"
	"time"

	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

  "github.com/mbta-performance-dashboard/consts"
  "github.com/mbta-performance-dashboard/utils"
)

func HandleCache(c *gin.Context, db *sql.DB, mutex *sync.Mutex) {
	fromStopIDs := strings.Split(c.DefaultQuery("from_stop_ids", ""), ",")
	toStopIDs := strings.Split(c.DefaultQuery("to_stop_ids", ""), ",")
	routeID := c.DefaultQuery("route_id", "")

  mutex.Lock()
  defer mutex.Unlock()

  utils.ValidateIDs(db, append(fromStopIDs, toStopIDs...), routeID)

  statement := fmt.Sprintf(
		"SELECT from_stop_id, to_stop_id, route_id, value AT TIME ZONE 'America/New_York' FROM "+
			"last_travel_time_cache_datetime WHERE from_stop_id IN (%s) AND to_stop_id IN (%s) AND "+
      "route_id = %s",
		utils.PgPlaceholders(0, len(fromStopIDs)),
		utils.PgPlaceholders(len(fromStopIDs), len(fromStopIDs)+len(toStopIDs)),
		utils.PgPlaceholders(len(fromStopIDs)+len(toStopIDs), len(fromStopIDs)+len(toStopIDs) + 1),
	)

  params := []any{}
	for i := 0; i < len(fromStopIDs); i++ {
		params = append(params, fromStopIDs[i])
	}
	for i := 0; i < len(toStopIDs); i++ {
		params = append(params, toStopIDs[i])
	}
	params = append(params, routeID)

  rows, err := db.Query(statement, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": fmt.Sprintf("Error querying last travel time cache datetimes: %v", err),
		})
    return
	}

	var datetimes map[string]map[string]LastCacheDatetime = make(
    map[string]map[string]LastCacheDatetime,
  )
	for rows.Next() {
		var datetime LastCacheDatetime
		err = rows.Scan(
			&datetime.FromStopID,
      &datetime.ToStopID,
			&datetime.RouteID,
			&datetime.Value,
		)
		if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{
        "data": fmt.Sprintf("Error scanning last travel time cache datetimes: %v", err),
      })
      return
		}
    if datetimes[datetime.FromStopID] == nil {
      datetimes[datetime.FromStopID] = make(map[string]LastCacheDatetime)
    }
    datetimes[datetime.FromStopID][datetime.ToStopID] = datetime
	}
	rows.Close()

  startOfToday, err := utils.StartOfToday()
  if err != nil {
		utils.PropagateToResponse(c, err)
		return
  }
	endOfYesterday := startOfToday.Add(-1 * time.Second)

	var wg sync.WaitGroup
	results := make(chan []TravelTime)
	error := false

	for i := 0; i < len(fromStopIDs); i++ {
    for j := 0; j < len(toStopIDs); j++ {
      wg.Add(1)

      go func(fromStopID string, toStopID string) {
        defer wg.Done()

        tx, err := db.Begin()
        if err != nil {
          log.Printf("Error beginning transaction: %v\n", err)
          error = true
          return
        }
        defer tx.Rollback()

        datetime, datetimeOk := datetimes[fromStopID][toStopID]
        if datetimeOk &&
          (datetime.Value.After(startOfToday) ||
            datetime.Value.Equal(startOfToday)) {
          log.Printf("Cache for (%s to %s) is already up to date\n", fromStopID, toStopID)
          error = true
          return
        }

        var startOfRange time.Time
        var endOfRange time.Time

        if !datetimeOk ||
          startOfToday.Sub(datetime.Value).Hours()/24 >= float64(consts.MaxDays) {
          startOfRange = startOfToday.AddDate(0, 0, -consts.MaxDays)
        } else {
          startOfRange = datetime.Value
        }

        if startOfRange.AddDate(0, 0, 7).Before(startOfToday) {
          endOfRange = startOfRange.AddDate(0, 0, 7).Add(-1 * time.Second)
        } else {
          endOfRange = endOfYesterday
        }

        client := http.Client{}

        for true {
          req, err := http.NewRequest("GET", fmt.Sprintf("%s/traveltimes", consts.ApiPerformance), nil)
          if err != nil {
            log.Printf("Error creating HTTP request to travel times endpoint: %v\n", err)
            error = true
            return
          }

          query := req.URL.Query()
          query.Add("api_key", os.Getenv("PERFORMANCE_API_KEY"))
          query.Add("format", "json")
          query.Add("from_stop", fromStopID)
          query.Add("to_stop", toStopID)
          query.Add("route", routeID)
          query.Add("from_datetime", strconv.FormatInt(startOfRange.Unix(), 10))
          query.Add("to_datetime", strconv.FormatInt(endOfRange.Unix(), 10))
          req.URL.RawQuery = query.Encode()

          res, err := client.Do(req)
          if err != nil {
            log.Printf("Error fetching travel times: %v\n", err)
            error = true
            return
          }

          body, err := io.ReadAll(res.Body)
          if err != nil {
            log.Printf("Error reading travel times response body: %v\n", err)
            error = true
            return
          }
          res.Body.Close()

          var apiRes ApiResponse
          json.Unmarshal(body, &apiRes)
          for i := 0; i < len(apiRes.TravelTimes); i++ {
            apiRes.TravelTimes[i].FromStopID = fromStopID
            apiRes.TravelTimes[i].ToStopID = toStopID
          }
          results <- apiRes.TravelTimes

          startOfRange = startOfRange.AddDate(0, 0, 7)
          if endOfRange.Equal(endOfYesterday) {
            break
          } else if endOfRange.AddDate(0, 0, 7).Before(startOfToday) {
            endOfRange = endOfRange.AddDate(0, 0, 7)
          } else {
            endOfRange = endOfYesterday
          }
        }

        var statement string
        if !datetimeOk {
          statement = "INSERT INTO last_travel_time_cache_datetime (from_stop_id, to_stop_id, "+
            "route_id, value) VALUES ($1, $2, $3, $4)"
          _, err = tx.Exec(statement, fromStopID, toStopID, routeID, startOfToday)
        } else {
          statement = "UPDATE last_travel_time_cache_datetime SET value = $1 WHERE from_stop_id "+
            "= $2 AND to_stop_id = $3 AND route_id = $4"
          _, err = tx.Exec(statement, startOfToday, fromStopID, toStopID, routeID)
        }
        if err != nil {
          log.Printf("Error updating last travel time cache datetime: %v\n", err)
          error = true
          return
        }

        statement = fmt.Sprintf(
          "DELETE FROM travel_time WHERE dep_dt < DATE_TRUNC('day', NOW() AT TIME ZONE "+
            "'America/New_York' - INTERVAL '%d days')",
          consts.MaxDays,
        )
        _, err = tx.Exec(statement)
        if err != nil {
          log.Printf("Error deleting old travel times: %v\n", err)
          error = true
          return
        }

        err = tx.Commit()
        if err != nil {
          log.Printf("Error committing transaction: %v\n", err)
          error = true
          return
        }
      }(fromStopIDs[i], toStopIDs[j])
    }
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var travelTimes []TravelTime = []TravelTime{}
	for result := range results {
		travelTimes = append(travelTimes, result...)
	}


	if len(travelTimes) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"data": "No travel times to insert",
		})
		return
	}

	statement = "INSERT INTO travel_time (from_stop_id, to_stop_id, route_id, direction, dep_dt, "+
    "arr_dt, travel_time_sec, benchmark_travel_time_sec) SELECT " +
    "unnest($1::text[]) AS from_stop_id, " +
    "unnest($2::text[]) AS to_stop_id, " +
    "unnest($3::text[]) AS route_id, " +
    "unnest($4::boolean[]) AS direction, " +
    "TO_TIMESTAMP(unnest($5::int[])) AS dep_dt, " +
    "TO_TIMESTAMP(unnest($6::int[])) AS arr_dt, " +
    "unnest($7::int[]) AS travel_time_sec, " +
    "unnest($8::int[]) AS benchmark_travel_time_sec"

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
			panic(fmt.Sprintf("Error converting travel time seconds to integer: %v", err))
		}
		convertedBenchmarkTravelTimeSec, err := strconv.Atoi(travelTimes[i].BenchmarkTravelTimeSec)
		if err != nil {
			panic(fmt.Sprintf("Error converting benchmark travel time seconds to integer: %v", err))
		}

    paramFromStopIDs = append(paramFromStopIDs, travelTimes[i].FromStopID)
    paramToStopIDs = append(paramToStopIDs, travelTimes[i].ToStopID)
    paramRouteIDs = append(paramRouteIDs, travelTimes[i].RouteID)
    paramDirections = append(paramDirections, convertedDirection == 1)
    paramDepDts = append(paramDepDts, convertedDepDt)
    paramArrDts = append(paramArrDts, convertedArrDt)
    paramTravelTimeSecs = append(paramTravelTimeSecs, convertedTravelTimeSec)
    paramBenchmarkTravelTimeSecs = append(
      paramBenchmarkTravelTimeSecs,
      convertedBenchmarkTravelTimeSec,
    )
	}

	_, err = db.Exec(
    statement,
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": fmt.Sprintf("Error inserting travel times: %v", err),
		})
		return
	}

	if error {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": "Cached new travel times, but error occurred - some may be missing",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": "Successfully cached new travel times",
	})
}

func HandleFetch(c *gin.Context, db *sql.DB) {
	fromStopIDs := strings.Split(c.DefaultQuery("from_stop_ids", ""), ",")
	toStopIDs := strings.Split(c.DefaultQuery("to_stop_ids", ""), ",")
	routeID := c.DefaultQuery("route_id", "")
	startDatetime := c.DefaultQuery("start_datetime", "")
	endDatetime := c.DefaultQuery("end_datetime", "")

	statement := "SELECT from_stop_id, to_stop_id, route_id, direction, dep_dt AT TIME ZONE "+
    "'America/New_York', arr_dt AT TIME ZONE 'America/New_York', travel_time_sec, "+
    "benchmark_travel_time_sec FROM travel_time WHERE 1=1"

  var params []any = []any{}

	if len(fromStopIDs) > 0 {
		statement += fmt.Sprintf(" AND from_stop_id IN (%s)", utils.PgPlaceholders(0, len(fromStopIDs)))
		for i := 0; i < len(fromStopIDs); i++ {
			params = append(params, fromStopIDs[i])
		}
	}

	if len(toStopIDs) > 0 {
		statement += fmt.Sprintf(
      " AND to_stop_id IN (%s)",
      utils.PgPlaceholders(len(fromStopIDs), len(fromStopIDs) + len(toStopIDs)),
    )
		for i := 0; i < len(toStopIDs); i++ {
			params = append(params, toStopIDs[i])
		}
	}

	if routeID != "" {
		statement += fmt.Sprintf(
      " AND route_id = %s",
      utils.PgPlaceholders(
        len(fromStopIDs) + len(toStopIDs),
        len(fromStopIDs) + len(toStopIDs) + 1,
      ),
    )
		params = append(params, routeID)
	}

	if startDatetime != "" {
		convertedStartDatetime, err := strconv.Atoi(startDatetime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"data": fmt.Sprintf("Error converting start datetime to integer: %v", err),
			})
			return
		}
		statement += fmt.Sprintf(
			" AND dep_dt >= TO_TIMESTAMP(%s)",
      utils.PgPlaceholders(
        len(fromStopIDs) + len(toStopIDs) + 1,
        len(fromStopIDs) + len(toStopIDs) + 2,
      ),
		)
		params = append(params, convertedStartDatetime)
	}

	if endDatetime != "" {
		convertedEndDatetime, err := strconv.Atoi(endDatetime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"data": fmt.Sprintf("Error converting end datetime to integer: %v", err),
			})
			return
		}
		statement += fmt.Sprintf(
			" AND dep_dt <= TO_TIMESTAMP(%s)",
      utils.PgPlaceholders(
        len(fromStopIDs) + len(toStopIDs) + 2,
        len(fromStopIDs) + len(toStopIDs) + 3,
      ),
		)
		params = append(params, convertedEndDatetime)
	}

	rows, err := db.Query(statement, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": fmt.Sprintf("Failed to fetch travel times: %v", err),
		})
		return
	}

	var travelTimes []TravelTime = []TravelTime{}
	for rows.Next() {
		var travelTime TravelTime
		err := rows.Scan(
			&travelTime.FromStopID,
      &travelTime.ToStopID,
			&travelTime.RouteID,
			&travelTime.Direction,
			&travelTime.DepDt,
			&travelTime.ArrDt,
			&travelTime.TravelTimeSec,
      &travelTime.BenchmarkTravelTimeSec,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"type": "error",
				"data": fmt.Sprintf("Error scanning travel times: %v", err),
			})
			return
		}
    travelTimes = append(travelTimes, travelTime)
	}
	rows.Close()

	c.JSON(http.StatusOK, gin.H{
		"data": travelTimes,
	})
}
