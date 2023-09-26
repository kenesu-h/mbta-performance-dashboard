-- migrate:up
CREATE TABLE IF NOT EXISTS route (
  id VARCHAR(255) PRIMARY KEY NOT NULL
);

CREATE TABLE IF NOT EXISTS shape (
  id VARCHAR(255) PRIMARY KEY NOT NULL,
  route_id VARCHAR(255) NOT NULL,
  polyline VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS stop (
  id VARCHAR(255) NOT NULL,
  route_id VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS last_headway_cache_datetime (
  stop_id VARCHAR(255) NOT NULL,
  route_id VARCHAR(255) NOT NULL,
  value TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS headway (
  stop_id VARCHAR(255) NOT NULL,
  route_id VARCHAR(255) NOT NULL,
  prev_route_id VARCHAR(255) NOT NULL,
  direction BOOLEAN NOT NULL,
  current_dep_dt TIMESTAMP NOT NULL,
  previous_dep_dt TIMESTAMP NOT NULL,
  headway_time_sec INT NOT NULL,
  benchmark_headway_time_sec INT NOT NULL
);

CREATE TABLE IF NOT EXISTS last_dwell_cache_datetime (
  stop_id VARCHAR(255) NOT NULL,
  route_id VARCHAR(255) NOT NULL,
  value TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS dwell (
  stop_id VARCHAR(255) NOT NULL,
  route_id VARCHAR(255) NOT NULL,
  direction BOOLEAN NOT NULL,
  arr_dt TIMESTAMP NOT NULL,
  dep_dt TIMESTAMP NOT NULL,
  dwell_time_sec INT NOT NULL
);

CREATE TABLE IF NOT EXISTS last_travel_time_cache_datetime (
  from_stop_id VARCHAR(255) NOT NULL,
  to_stop_id VARCHAR(255) NOT NULL,
  route_id VARCHAR(255) NOT NULL,
  value TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS travel_time (
  from_stop_id VARCHAR(255) NOT NULL,
  to_stop_id VARCHAR(255) NOT NULL,
  route_id VARCHAR(255) NOT NULL,
  direction BOOLEAN NOT NULL,
  dep_dt TIMESTAMP NOT NULL,
  arr_dt TIMESTAMP NOT NULL,
  travel_time_sec INT NOT NULL,
  benchmark_travel_time_sec INT NOT NULL
);

-- migrate:down
DROP TABLE route;

DROP TABLE shape;

DROP TABLE stop;

DROP TABLE last_headway_cache_datetime;

DROP TABLE headway;

DROP TABLE last_dwell_cache_datetime;

DROP TABLE dwell;

DROP TABLE last_travel_time_cache_datetime;

DROP TABLE travel_time;
