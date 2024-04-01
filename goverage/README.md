# Goverage Service

A coverage API service that allows you to publish coverage data and retrieve coverage data.

## Usage

1. Deploy the service to a server, along with a PostgreSQL database.
2. Configure the `goverage_host` and `goverage_token` inputs in the action to use it.

## Configuration

The service requires the following environment variables to be set:

- `GOVERAGE_DB_CONN_STR`: Connection string in the form `user=USER password=PASSWORD dbname=DBNAME host=HOST port=PORT sslmode=disable`
- `GOVERAGE_API_KEY`: Secret key for the service

The service will listen on port `1323`.
