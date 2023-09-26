package headways

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
	_ "github.com/lib/pq"

  "github.com/mbta-performance-dashboard/consts"
  "github.com/mbta-performance-dashboard/types"
  "github.com/mbta-performance-dashboard/utils"
)

func HandleCache(c *gin.Context, db *sql.DB, mutex *sync.Mutex) {
	stopIDs := strings.Split(c.DefaultQuery("stop_ids", ""), ",")
	routeID := c.DefaultQuery("route_id", "")

  mutex.Lock()
  defer mutex.Unlock()

  utils.ValidateIDs(db, stopIDs, routeID)

  statement := fmt.Sprintf(
		"SELECT stop_id, route_id, value AT TIME ZONE 'America/New_York' FROM "+
			"last_headway_cache_datetime WHERE stop_id IN (%s) AND route_id = %s",
		utils.PgPlaceholders(0, len(stopIDs)),
		utils.PgPlaceholders(len(stopIDs), len(stopIDs)+1),
	)

  params := []any{}
	for i := 0; i < len(stopIDs); i++ {
		params = append(params, stopIDs[i])
	}
	params = append(params, routeID)

  rows, err := db.Query(statement, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": fmt.Sprintf("Error querying last headway cache datetimes: %v", err),
		})
    return
	}

	var datetimes map[string]types.LastCacheDatetime = make(map[string]types.LastCacheDatetime)
	for rows.Next() {
		var datetime types.LastCacheDatetime
		err = rows.Scan(
			&datetime.StopID,
			&datetime.RouteID,
			&datetime.Value,
		)
		if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{
        "data": fmt.Sprintf("Error scanning last headway cache datetimes: %v", err),
      })
      return
		}
		datetimes[datetime.StopID] = datetime
	}
	rows.Close()

  startOfToday, err := utils.StartOfToday()
  if err != nil {
		utils.PropagateToResponse(c, err)
		return
  }
	endOfYesterday := startOfToday.Add(-1 * time.Second)

	var wg sync.WaitGroup
	results := make(chan []Headway)
	error := false

	for i := 0; i < len(stopIDs); i++ {
		wg.Add(1)

		go func(stopID string) {
			defer wg.Done()

			tx, err := db.Begin()
			if err != nil {
				log.Printf("Error beginning transaction: %v\n", err)
				error = true
				return
			}
			defer tx.Rollback()

			datetime, datetimeOk := datetimes[stopID]
			if datetimeOk &&
				(datetime.Value.After(startOfToday) ||
					datetime.Value.Equal(startOfToday)) {
				log.Printf("Cache for %s is already up to date\n", stopID)
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
				req, err := http.NewRequest("GET", fmt.Sprintf("%s/headways", consts.ApiPerformance), nil)
				if err != nil {
					log.Printf("Error creating HTTP request to headways endpoint: %v\n", err)
					error = true
					return
				}

				query := req.URL.Query()
				query.Add("api_key", os.Getenv("PERFORMANCE_API_KEY"))
				query.Add("format", "json")
				query.Add("stop", stopID)
				query.Add("route", routeID)
				query.Add("from_datetime", strconv.FormatInt(startOfRange.Unix(), 10))
				query.Add("to_datetime", strconv.FormatInt(endOfRange.Unix(), 10))
				req.URL.RawQuery = query.Encode()

				res, err := client.Do(req)
				if err != nil {
					log.Printf("Error fetching headways: %v\n", err)
					error = true
					return
				}

				body, err := io.ReadAll(res.Body)
				if err != nil {
					log.Printf("Error reading headways response body: %v\n", err)
					error = true
					return
				}
				res.Body.Close()

				var apiRes ApiResponse
				json.Unmarshal(body, &apiRes)
				for i := 0; i < len(apiRes.Headways); i++ {
					apiRes.Headways[i].StopID = stopID
				}
				results <- apiRes.Headways

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
				statement = "INSERT INTO last_headway_cache_datetime (stop_id, route_id, value) VALUES " +
					"($1, $2, $3)"
				_, err = tx.Exec(statement, stopID, routeID, startOfToday)
			} else {
				statement = "UPDATE last_headway_cache_datetime SET value = $1 WHERE stop_id = $2 AND " +
					"route_id = $3"
				_, err = tx.Exec(statement, startOfToday, stopID, routeID)
			}
			if err != nil {
				log.Printf("Error updating last headway cache datetime: %v\n", err)
				error = true
				return
			}

			statement = fmt.Sprintf(
				"DELETE FROM headway WHERE current_dep_dt < DATE_TRUNC('day', NOW() AT TIME ZONE "+
					"'America/New_York' - INTERVAL '%d days')",
				consts.MaxDays,
			)
			_, err = tx.Exec(statement)
			if err != nil {
				log.Printf("Error deleting old headways: %v\n", err)
				error = true
				return
			}

			err = tx.Commit()
			if err != nil {
				log.Printf("Error committing transaction: %v\n", err)
				error = true
				return
			}
		}(stopIDs[i])
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var headways []Headway = []Headway{}
	for result := range results {
		headways = append(headways, result...)
	}


	if len(headways) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"data": "No headways to insert",
		})
		return
	}

	statement = "INSERT INTO headway (stop_id, route_id, prev_route_id, direction, " +
		"current_dep_dt, previous_dep_dt, headway_time_sec, benchmark_headway_time_sec) VALUES"

	params = []any{}
	var values []string = []string{}
	for i := 0; i < len(headways); i++ {
		start := (i * 8) + 1
		values = append(
			values,
			fmt.Sprintf(
				" ($%d, $%d, $%d, $%d, TO_TIMESTAMP($%d), TO_TIMESTAMP($%d), $%d, $%d)",
				start,
				start+1,
				start+2,
				start+3,
				start+4,
				start+5,
				start+6,
				start+7,
			),
		)

		convertedDirection, err := strconv.Atoi(headways[i].Direction)
		if err != nil {
			panic(fmt.Sprintf("Error converting direction to integer: %v", err))
		}
		convertedCurrentDepDt, err := strconv.Atoi(headways[i].CurrentDepDt)
		if err != nil {
			panic(fmt.Sprintf("Error converting current departure datetime to integer: %v", err))
		}
		convertedPreviousDepDt, err := strconv.Atoi(headways[i].PreviousDepDt)
		if err != nil {
			panic(fmt.Sprintf("Error converting previous departure datetime to integer: %v", err))
		}
		convertedHeadwayTimeSec, err := strconv.Atoi(headways[i].HeadwayTimeSec)
		if err != nil {
			panic(fmt.Sprintf("Error converting headway time seconds to integer: %v", err))
		}
		convertedBenchmarkHeadwayTimeSec, err := strconv.Atoi(headways[i].BenchmarkHeadwayTimeSec)
		if err != nil {
			panic(fmt.Sprintf("Error converting benchmark headway time seconds to integer: %v", err))
		}

		params = append(
			params,
			[]any{
				headways[i].StopID,
				headways[i].RouteID,
				headways[i].PrevRouteID,
				convertedDirection == 1,
				convertedCurrentDepDt,
				convertedPreviousDepDt,
				convertedHeadwayTimeSec,
				convertedBenchmarkHeadwayTimeSec,
			}...,
		)
	}
	statement += strings.Join(values, ",")

	_, err = db.Exec(statement, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": fmt.Sprintf("Error inserting headways: %v", err),
		})
		return
	}

	if error {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": "Cached new headways, but error occurred - some may be missing",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": "Successfully cached new headways",
	})
}

func HandleFetch(c *gin.Context, db *sql.DB) {
	stopIDs := strings.Split(c.DefaultQuery("stop_ids", ""), ",")
	routeID := c.DefaultQuery("route_id", "")
	startDatetime := c.DefaultQuery("start_datetime", "")
	endDatetime := c.DefaultQuery("end_datetime", "")

	statement := "SELECT stop_id, route_id, prev_route_id, direction, current_dep_dt AT TIME " +
		"ZONE 'America/New_York', previous_dep_dt AT TIME ZONE 'America/New_York', " +
		"headway_time_sec, benchmark_headway_time_sec FROM headway WHERE 1=1"

	var params []any = []any{}

	if len(stopIDs) > 0 {
		statement += fmt.Sprintf(" AND stop_id IN (%s)", utils.PgPlaceholders(0, len(stopIDs)))
		for i := 0; i < len(stopIDs); i++ {
			params = append(params, stopIDs[i])
		}
	}

	if routeID != "" {
		statement += fmt.Sprintf(" AND route_id = %s", utils.PgPlaceholders(len(stopIDs), len(stopIDs)+1))
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
			" AND current_dep_dt >= TO_TIMESTAMP(%s)",
			utils.PgPlaceholders(len(stopIDs)+1, len(stopIDs)+2),
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
			" AND current_dep_dt <= TO_TIMESTAMP(%s)",
			utils.PgPlaceholders(len(stopIDs)+2, len(stopIDs)+3),
		)
		params = append(params, convertedEndDatetime)
	}

	rows, err := db.Query(statement, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": fmt.Sprintf("Failed to fetch headways: %v", err),
		})
		return
	}

	var headways []Headway = []Headway{}
	for rows.Next() {
		var headway Headway
		err := rows.Scan(
			&headway.StopID,
			&headway.RouteID,
			&headway.PrevRouteID,
			&headway.Direction,
			&headway.CurrentDepDt,
			&headway.PreviousDepDt,
			&headway.HeadwayTimeSec,
			&headway.BenchmarkHeadwayTimeSec,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"data": fmt.Sprintf("Error scanning headways: %v", err),
			})
			return
		}
		headways = append(headways, headway)
	}
	rows.Close()

	c.JSON(http.StatusOK, gin.H{
		"data": headways,
	})
}
