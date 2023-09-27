#!/bin/bash

POSTGRES_URL="postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable"

until psql "$POSTGRES_URL" -c '\q'; do
  sleep 1
done

echo "Running database migrations"
dbmate -u "$POSTGRES_URL" -d "/app/db/migrations" up

echo "Running initial cache"
go run ./cache

echo "Compiling to binary"
go build -o main .

echo "Running binary"
./main
