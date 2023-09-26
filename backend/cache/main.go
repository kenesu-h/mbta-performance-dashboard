package main

import (
	"fmt"
	"io"
  "log"
	"os"
	"strings"

  "database/sql"
	"encoding/json"
	"net/http"

	"github.com/joho/godotenv"
  _ "github.com/lib/pq"
)

type RoutesResponse struct {
  Included []RoutePattern `json:"included"`
}

type RoutePattern struct {
  Relationship Relationship `json:"relationships"`
}

type Relationship struct {
  RepresentativeTrip RepresentativeTrip `json:"representative_trip"`
}

type RepresentativeTrip struct {
  Data RepresentativeTripData `json:"data"`
}

type RepresentativeTripData struct {
  ID string `json:"id"`
}

type TripsResponse struct {
  Included []Entity `json:"included"`
}

type Entity struct {
  ID         string          `json:"id"`
  Type       string          `json:"type"`
  Attributes json.RawMessage `json:"attributes"`
}

type ShapeAttributes struct {
  Polyline string `json:"polyline"`
}

type StopAttributes struct {
  Latitude  float64 `json:"latitude"`
  Longitude float64 `json:"longitude"`
  Name      string  `json:"name"`
}

type Route struct {
  ID string `json:"id"`
}

type Shape struct {
  ID          string      `json:"id"`
  RouteID     string      `json:"route_id"`
  Polyline    string      `json:"polyline"`
}

type Stop struct {
  ID         string    `json:"id"`
  RouteID    string    `json:"route_id"`
  Name       string    `json:"name"`
  Latitude   float64   `json:"latitude"`
  Longitude  float64   `json:"longitude"`
}

const V3Api string = "https://api-v3.mbta.com"

func routeIDs() []string {
  return []string {
    "Red",
    "Mattapan",
    "Orange",
    "Green-B",
    "Green-C",
    "Green-D",
    "Green-E",
    "Blue",
  }
}

func main() {
  err := godotenv.Load()
  if err != nil {
    panic(fmt.Sprintf("Error loading .env file: %v", err))
  }
  apiKey, apiKeyExists := os.LookupEnv("V3_API_KEY")

  client := http.Client{}
  var routes []Route
  var shapes []Shape
  var stops []Stop
  for _, routeID := range routeIDs() {
    req, err := http.NewRequest("GET", fmt.Sprintf("%s/routes", V3Api), nil)
    if err != nil {
      panic(fmt.Sprintf("Error creating HTTP request to routes endpoint: %v", err))
    }

    if apiKeyExists {
      req.Header.Add("x-api-key", apiKey)
    }
    query := req.URL.Query()
    query.Add("fields[route]", "id")
    query.Add("include", "route_patterns")
    query.Add("filter[id]", routeID)
    req.URL.RawQuery = query.Encode()

    log.Println(fmt.Sprintf("Fetching route %s", routeID))
    res, err := client.Do(req)
    if err != nil {
      panic(fmt.Sprintf("Error fetching route %s: %v", routeID, err))
    }

    body, err := io.ReadAll(res.Body)
    if err != nil {
      panic(fmt.Sprintf("Error reading routes response body: %v", err))
    }
    res.Body.Close()

    var tripIDs []string
    var routesRes RoutesResponse
    json.Unmarshal(body, &routesRes)
    for _, routePattern := range routesRes.Included {
      tripIDs = append(tripIDs, routePattern.Relationship.RepresentativeTrip.Data.ID)
    }

    log.Println("Fetching accumulated trips")
    req, err = http.NewRequest("GET", fmt.Sprintf("%s/trips", V3Api), nil)
    if err != nil {
      panic(fmt.Sprintf("Error creating HTTP request to trips endpoint: %v", err))
    }
    if apiKeyExists {
      req.Header.Add("x-api-key", apiKey)
    }
    query = req.URL.Query()
    query.Add("include", "shape,stops")
    query.Add("filter[id]", strings.Join(tripIDs[:], ","))
    req.URL.RawQuery = query.Encode()

    res, err = client.Do(req)
    if err != nil {
      panic(fmt.Sprintf("Error fetching trips: %v", err))
    }

    body, err = io.ReadAll(res.Body)
    if err != nil {
      panic(fmt.Sprintf("Error reading trips response body: %v", err))
    }
    res.Body.Close()

    routes = append(routes, Route {
      ID: routeID,
    })

    var tripsRes TripsResponse
    json.Unmarshal(body, &tripsRes)
    for _, entity := range tripsRes.Included {
      switch entity.Type {
      case "shape":
        var shapeAttr ShapeAttributes
        json.Unmarshal(entity.Attributes, &shapeAttr);
        shapes = append(shapes, Shape {
          ID: entity.ID,
          RouteID: routeID,
          Polyline: shapeAttr.Polyline,
        })
      case "stop":
        var stopAttr StopAttributes
        json.Unmarshal(entity.Attributes, &stopAttr)
        stops = append(stops, Stop {
          ID: entity.ID,
          RouteID: routeID,
          Name: stopAttr.Name,
          Latitude: stopAttr.Latitude,
          Longitude: stopAttr.Longitude,
        })
      }
    }
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

  log.Println("Clearing existing cache")
  _, err = db.Exec("DELETE FROM route")
  if err != nil {
    panic(fmt.Sprintf("Error clearing route table: %v", err))
  }

  _, err = db.Exec("DELETE FROM shape")
  if err != nil {
    panic(fmt.Sprintf("Error clearing shape table: %v", err))
  }

  _, err = db.Exec("DELETE FROM stop")
  if err != nil {
    panic(fmt.Sprintf("Error clearing stop table: %v", err))
  }

  if len(routes) > 0 {
    log.Println("Inserting routes into cache")
    statement := "INSERT INTO route (id) VALUES "
    var values []string
    for _, route := range routes {
      values = append(
        values,
        fmt.Sprintf("('%s')", route.ID),
      )
    }
    statement += strings.Join(values, ", ")

    prepared, err := db.Prepare(statement)
    if err != nil {
      panic(fmt.Sprintf("Error preparing routes statement: %v", err))
    }
    _, err = prepared.Exec()
    if err != nil {
      panic(fmt.Sprintf("Error inserting routes: %v", err))
    }
  } else {
    log.Println("No routes to insert")
  }

  if len(shapes) > 0 {
    log.Println("Inserting shapes into cache")
    statement := "INSERT INTO shape (id, route_id, polyline) VALUES "
    var values []string
    for _, shape := range shapes {
      values = append(
        values,
        fmt.Sprintf("('%s', '%s', '%s')", shape.ID, shape.RouteID, shape.Polyline),
      )
    }
    statement += strings.Join(values, ", ")

    prepared, err := db.Prepare(statement)
    if err != nil {
      panic(fmt.Sprintf("Error preparing shapes statement: %v", err))
    }
    _, err = prepared.Exec()
    if err != nil {
      panic(fmt.Sprintf("Error inserting shapes: %v", err))
    }
  } else {
    log.Println("No shapes to insert")
  }

  if len(stops) > 0 {
    log.Println("Inserting stops into cache")
    statement := "INSERT INTO stop (id, route_id, name, latitude, longitude) VALUES "
    var values []string
    for _, stop := range stops {
      values = append(
        values,
        fmt.Sprintf(
          "('%s', '%s', $$%s$$, %f, %f)",
          stop.ID,
          stop.RouteID,
          stop.Name,
          stop.Latitude,
          stop.Longitude,
        ),
      )
    }
    statement += strings.Join(values, ", ")

    prepared, err := db.Prepare(statement)
    if err != nil {
      panic(fmt.Sprintf("Error preparing stops statement: %v", err))
    }
    _, err = prepared.Exec() 
    if err != nil {
      panic(fmt.Sprintf("Error inserting stops: %v", err))
    }
  } else {
    log.Println("No stops to insert")
  }

  log.Println("Done!")
}
