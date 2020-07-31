.EXPORT_ALL_VARIABLES:

COMPOSE_CONVERT_WINDOWS_PATHS=1

BUILD_ROTATOR_PATH = ./cmd/rotator/main.go
BUILD_ROTATOR = rotator
CONFIG_ROTATOR_PATH = cmd/rotator/config/config.yaml

BUILD_AGGREGATOR_PATH = ./cmd/aggregator/main.go
BUILD_AGGREGATOR = aggregator
CONFIG_AGGREGATOR_PATH = cmd/aggregator/config/config.yaml

KAFKA_TOPIC = topic_rotator
KAFKA_ADDR = localhost:9092

PG_SCHEMAPATH = ./internal/data/repository/sql
DSN = host=localhost port=5432 user=igor password=igor dbname=rotator sslmode=disable

GRPC_PORT = 4445
GRPS_GW_PORT = 4446

COMMAND_START = ./rotator --config config.yaml run

initdb:
COMMAND_START = ./rotator --config config.yaml run --updb
docker_env:
DSN	= host=db port=5432 user=igor password=igor dbname=rotator sslmode=disable
KAFKA_ADDR = kafka:9092

os:
ifeq ($(OS),Windows_NT)
PG_SCHEMAPATH = .\internal\data\repository\sql
BUILD_ROTATOR_PATH =  cmd\rotator\main.go
CONFIG_ROTATOR_PATH = cmd\rotator\config\config.yaml
BUILD_AGGREGATOR_PATH = cmd\aggregator\main.go
CONFIG_AGGREGATOR_PATH = cmd\aggregator\config\config.yaml
BUILD_ROTATOR = rotator.exe
BUILD_AGGREGATOR = aggregator.exe
endif

tidy:
	go mod tidy
fmt:
	go fmt ./...
prepare_lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0
lint: prepare_lint fmt tidy
	golangci-lint run ./...
run_rotator:
	go run cmd/rotator/main.go --config $(CONFIG_ROTATOR_PATH) --debug run --updb
run_aggregator:
	go run cmd/aggregator/main.go --config $(CONFIG_AGGREGATOR_PATH) --debug run
build:
	go build -o $(BUILD_ROTATOR) $(BUILD_ROTATOR_PATH)
	go build -o $(BUILD_AGGREGATOR) $(BUILD_AGGREGATOR_PATH)
test:
	go test -race ./...
testv:
	go test -v -race ./...
testi:
	go test -p 1 -v ./tests -tags=integration
prepare_gen:
		go get \
            github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
            github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
            github.com/golang/protobuf/protoc-gen-go
gen:
	protoc -I cmd/rotator/api/ api.proto --go_out=plugins=grpc:cmd/rotator/api --grpc-gateway_out=logtostderr=true:cmd/rotator/api

db-up-dev:
	$(BUILD_ROTATOR) --config $(CONFIG_ROTATOR_PATH) run --updb

docker-up-dev: os
	docker-compose -f docker-compose-dev.yml up -d
docker-down-dev:
	docker-compose -f docker-compose-dev.yml down -v --remove-orphans
kill-dev:
	taskkill /IM $(BUILD_ROTATOR) 2>nul
	taskkill /IM $(BUILD_AGGREGATOR) 2>nul

up-dev: prepare_gen gen docker-up-dev build db-up-dev
	$(BUILD_ROTATOR) --config $(CONFIG_ROTATOR_PATH) --debug run
down-dev: docker-down-dev

up-with-db: initdb docker_env prepare_gen gen docker-up
down-with-db: docker-down-hard

up:  docker_env prepare_gen gen docker-up
down: docker-down

integration-test: up-with-db testi down-with-db

docker-up: os
	docker-compose -f docker-compose.yml up -d
docker-down:
	docker-compose -f docker-compose.yml down
docker-down-hard:
	docker-compose -f docker-compose.yml down -v --remove-orphans
	docker rmi banner_rotator_rotator:latest
	docker rmi banner_rotator_aggregator:latest
.PHONY: build, all, fmt, lint, test, run, os,up,down,gen,tidy, generate