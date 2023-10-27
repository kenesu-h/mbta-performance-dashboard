package traveltimes

import (
  "errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
  "github.com/mbta-performance-dashboard/utils"
)

func CacheTravelTimes(c *gin.Context, service *TravelTimeService) {
	fromStopIDs := strings.Split(c.DefaultQuery("from_stop_ids", ""), ",")
	toStopIDs := strings.Split(c.DefaultQuery("to_stop_ids", ""), ",")
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

    if err := utils.ValidateIDs(tx, append(fromStopIDs, toStopIDs...), routeID); err != nil {
      return err
    }

    entities, errs := service.FetchTravelTimesFromAPI(tx, fromStopIDs, toStopIDs, routeID)
    if len(errs) > 0 {
      return errors.Join(errs...)
    }

    if err = service.Insert(tx, entities); err != nil {
      return err
    }

    if err = service.UpdateTravelTimeCacheDatetimes(tx, fromStopIDs, toStopIDs, routeID); err != nil {
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
    utils.PropagateToResponse(c, err)
    return
  }

	c.JSON(http.StatusOK, gin.H{
		"data": "Successfully cached new entities",
	})
}

func SelectTravelTimes(c *gin.Context, service *TravelTimeService) {
	fromStopIDs := strings.Split(c.DefaultQuery("from_stop_ids", ""), ",")
	toStopIDs := strings.Split(c.DefaultQuery("to_stop_ids", ""), ",")
	routeID := c.DefaultQuery("route_id", "")

  var travelTimes []*TravelTime
  err := func() error {
    tx, err := service.BeginTx()
    if err != nil {
      return fmt.Errorf("Error beginning transaction: %w", err)
    }

    if err := utils.ValidateIDs(tx, append(fromStopIDs, toStopIDs...), routeID); err != nil {
      return err
    }

    travelTimes, err = service.SelectTravelTimes(tx, fromStopIDs, toStopIDs, routeID)
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
    utils.PropagateToResponse(c, err)
    return
  }

	c.JSON(http.StatusOK, gin.H{
		"data": travelTimes,
	})
}
