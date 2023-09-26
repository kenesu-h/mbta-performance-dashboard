package utils

import (
  "errors"
	"fmt"
	"strings"
  "time"

  "database/sql"
  "net/http"

  "github.com/gin-gonic/gin"
 
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

// ValidateIDs validates the provided stop and route IDs. Returns an error if at least one is
// invalid.
func ValidateIDs(db *sql.DB, stopIDs []string, routeID string) error {
	if len(stopIDs) == 0 {
		return errors.New("At least one stop ID required")
	}

	if routeID == "" {
		return errors.New("Route ID required")
	}

  statement := fmt.Sprintf("SELECT * FROM stop WHERE id IN (%s)", PgPlaceholders(0, len(stopIDs)))

	var params []any = []any{}
	for i := 0; i < len(stopIDs); i++ {
		params = append(params, stopIDs[i])
	}

	rows, err := db.Query(statement, params...)
	if err != nil {
		return errors.New(fmt.Sprintf("Error querying stops: %v", err))
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
			return errors.New(fmt.Sprintf("Error scanning stops: %v", err))
		}
		stops = append(stops, stop)
	}
	rows.Close()

	if len(stops) != len(stopIDs) {
		return errors.New("Not all stop IDs are valid")
	}

  statement = "SELECT * from route WHERE id = $1"

  rows, err = db.Query(statement, routeID)
	if err != nil {
		return errors.New(fmt.Sprintf("Error querying routes: %v", err))
	}

	if !rows.Next() {
		return errors.New(fmt.Sprintf("Invalid route ID %s", routeID))
	}
	rows.Close()

  return nil
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
