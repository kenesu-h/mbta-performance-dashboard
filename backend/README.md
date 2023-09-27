# Backend

## Prerequisites

- [Docker Engine](https://docs.docker.com/engine/)
- [Docker Compose](https://docs.docker.com/compose/)

## Quick Setup

1. Create a .env containing the following variables:
    - PERFORMANCE_API_KEY: Required
        - Request a key by emailing [developer@mbta.com](mailto:developer@mbta.com).
        - If you don't have a key, use [the MBTA's open development key]
          (https://cdn.mbta.com/sites/default/files/2017-11/api-public-key.txt)
        - Rate limits for the performance API are unknown.
    - V3_API_KEY: Optional
        - Request a key at <https://api-v3.mbta.com/>
        - Without a key (as of 9/7/2023), [you are limited to 20 requests per minute](
          https://www.mbta.com/developers/v3-api/best-practices).
2. `docker-compose up -d dev`

## Caching

The MBTA's Performance API only allows you to query up to 90 days worth of data, while restricting
each query to the timespan of a week or less.

To (try to) avoid the ire of the MBTA, this backend will only cache up to the last 30 days worth of
data and will delete all data older than that when an endpoint is hit. This is also intended to
keep the cache at a manageable size.

Feel free to remove the lines that delete the old data, but do so at your own risk.
