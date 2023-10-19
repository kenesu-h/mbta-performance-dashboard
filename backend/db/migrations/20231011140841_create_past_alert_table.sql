-- migrate:up
CREATE TABLE IF NOT EXISTS last_past_alert_cache_datetime (
  stop_id VARCHAR(255) NOT NULL,
  route_id VARCHAR(255) NOT NULL,
  value TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS past_alert (
  id INT PRIMARY KEY NOT NULL
);

CREATE TABLE IF NOT EXISTS alert_version (
  alert_id INT NOT NULL,
  version_id INT NOT NULL,
  valid_from TIMESTAMP NOT NULL,
  valid_to TIMESTAMP NOT NULL,
  cause VARCHAR(255) NOT NULL,
  effect VARCHAR(255) NOT NULL,
  header_text VARCHAR(255) NOT NULL,
  description_text VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS informed_entity (
  alert_id INT NOT NULL,
  version_id INT NOT NULL,
  agency_id VARCHAR(255),
  route_id VARCHAR(255),
  route_type INT,
  trip_id VARCHAR(255),
  stop_id VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS active_period (
  alert_id INT NOT NULL,
  version_id INT NOT NULL,
  start TIMESTAMP NOT NULL,
  end TIMESTAMP
);

-- migrate:down
DROP TABLE last_past_alert_cache_datetime;

DROP TABLE past_alert;

DROP TABLE informed_entity;

DROP TABLE active_period;
