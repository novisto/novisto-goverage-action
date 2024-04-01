default:
	@echo "See usage in Makefile"

.PHONY: install run lint tests build run-docker kill-docker sqlc docker-build docker-run docker-kill start-pg stop-pg

LOCAL_DB_CONN_STR="user=root dbname=root password=root host=127.0.0.1 port=54350 sslmode=disable"

install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod download

run:
	GOVERAGE_DB_CONN_STR=$(LOCAL_DB_CONN_STR) GOVERAGE_API_KEY=valid-key go run ./server.go

lint:
	go vet
	golangci-lint run

tests:
	go test -v .

sqlc:
	rm -rf data/*
	sqlc generate

build: sqlc
	go build .

docker-build: sqlc
	docker buildx build -t goverage-api:latest .

docker-run: docker-build
	docker run -d --rm -p 1324:1324 --name goverage-api goverage-api:latest

docker-kill:
	docker kill goverage-api

goose-up:
	goose -dir sql/versions postgres $(LOCAL_DB_CONN_STR) up

goose-down:
	goose -dir sql/versions postgres $(LOCAL_DB_CONN_STR) down

goose-create:
	goose -dir sql/versions create $(name) sql

pg-start:
	docker run -p 54350:5432 --name goverage-pg --rm -d \
		-e POSTGRES_PASSWORD=root \
		-e POSTGRES_USER=root \
		-e POSTGRES_DB=root \
		postgres:15

pg-stop:
	docker kill goverage-pg
