package main

import (
	"fmt"
	"os"
	"sync"

	"database/sql"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/mbta-performance-dashboard/dwells"
	"github.com/mbta-performance-dashboard/headways"
	"github.com/mbta-performance-dashboard/traveltimes"
	"github.com/mbta-performance-dashboard/types"
	"github.com/mbta-performance-dashboard/utils"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file: %v", err))
	}

	source := fmt.Sprintf(
		"host=%s port=%s dbname=%s password=%s user=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_USER"),
	)
	db, err := sql.Open("postgres", source)
	if err != nil {
		panic(fmt.Sprintf("Error opening database: %v", err))
	}
	defer db.Close()

	var mutex sync.Mutex

	r := gin.Default()
	r.Use(cors.Default())

	// shape : []Shape
	r.GET("/shape", func(c *gin.Context) {
		statement := "SELECT * FROM shape"

		prepared, err := db.Prepare(statement)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"type": "error",
				"data": fmt.Sprintf("Error preparing shapes statement: %v", err),
			})
			return
		}

		rows, err := prepared.Query()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"type": "error",
				"data": fmt.Sprintf("Failed to fetch shapes: %v", err),
			})
			return
		}

		var shapes []types.Shape = []types.Shape{}
		for rows.Next() {
			var shape types.Shape
			err := rows.Scan(&shape.ID, &shape.RouteID, &shape.Polyline)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"type": "error",
					"data": fmt.Sprintf("Error scanning shapes: %v", err),
				})
				return
			}
			shapes = append(shapes, shape)
		}
		rows.Close()

		c.JSON(http.StatusOK, gin.H{
			"type": "success",
			"data": shapes,
		})
	})

	// stop : -> []Stop
	r.GET("/stop", func(c *gin.Context) {
		statement := "SELECT * FROM stop"

		prepared, err := db.Prepare(statement)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"type": "error",
				"data": fmt.Sprintf("Error preparing stops statement: %v", err),
			})
			return
		}

		rows, err := prepared.Query()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"type": "error",
				"data": fmt.Sprintf("Failed to fetch stops: %v", err),
			})
			return
		}

		var stops []types.Stop = []types.Stop{}
		for rows.Next() {
			var stop types.Stop
			err := rows.Scan(&stop.ID, &stop.RouteID, &stop.Name, &stop.Latitude, &stop.Longitude)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"type": "error",
					"data": fmt.Sprintf("Error scanning stops: %v", err),
				})
				return
			}
			stops = append(stops, stop)
		}
		rows.Close()

		c.JSON(http.StatusOK, gin.H{
			"type": "success",
			"data": stops,
		})
	})

	headwayService := headways.NewService(db, &mutex)
	// /cache/headway : stop_ids []string, route_id string
	r.GET("/cache/headway", func(c *gin.Context) {
		utils.Cache[*headways.Headway](c, headwayService)
	})

	// /headway : stop_ids []string, route_id string, start_datetime int, end_datetime int -> []Headway
	r.GET("/headway", func(c *gin.Context) {
		utils.Select[*headways.Headway](c, headwayService)
	})

	dwellService := dwells.NewService(db, &mutex)
	// /cache/dwell : stop_ids []string, route_id string
	r.GET("/cache/dwell", func(c *gin.Context) {
		utils.Cache[*dwells.Dwell](c, dwellService)
	})

	// /dwell : stop_ids []string, route_id string, start_datetime int, end_datetime int -> []Dwell
	r.GET("/dwell", func(c *gin.Context) {
		utils.Select[*dwells.Dwell](c, dwellService)
	})

	travelTimeService := traveltimes.NewService(db, &mutex)
	// /cache/travel_time : from_stop_ids []string, to_stop_ids []string, route_id string
	r.GET("/cache/travel_time", func(c *gin.Context) {
		traveltimes.CacheTravelTimes(c, travelTimeService)
	})

	// /travel_time : from_stop_ids []string, to_stop_ids []string, route_id string -> []TravelTime
	r.GET("/travel_time", func(c *gin.Context) {
		traveltimes.SelectTravelTimes(c, travelTimeService)
	})

	r.Run()
}
