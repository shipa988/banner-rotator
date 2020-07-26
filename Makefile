.EXPORT_ALL_VARIABLES:

PG_SCHEMAPATH = ./internal/data/repository/sql
BUILD_ROTATOR_PATH = ./cmd/rotator/main.go
CONFIG_ROTATOR_PATH = cmd/rotator/config/config.yaml
BUILD_AGGREGATOR_PATH = ./cmd/aggregator/main.go
CONFIG_AGGREGATOR_PATH = cmd/aggregator/config/config.yaml
TOPIC = topic_rotator
BUILD_ROTATOR = rotator
BUILD_AGGREGATOR = aggregator
COMPOSE_CONVERT_WINDOWS_PATHS=1

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
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.29.0
lint: prepare_lint fmt tidy
	golangci-lint run ./...
run_rotator:
	go run cmd/rotator/main.go --config $(CONFIG_ROTATOR_PATH) --debug run
run_aggregator:
	go run cmd/aggregator/main.go --config $(CONFIG_AGGREGATOR_PATH) --debug run
build:
	go build -o $(BUILD_ROTATOR) $(BUILD_ROTATOR_PATH)
test:
	go test -race ./...
testv:
	go test -v -race ./...

prepare_gen:
		go get \
            github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
            github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
            github.com/golang/protobuf/protoc-gen-go
gen:
	protoc -I cmd/rotator/api/ api.proto --go_out=plugins=grpc:internal --grpc-gateway_out=logtostderr=true:cmd/rotator/internal

up: prepare_gen gen docker-up build
	$(BUILD_ROTATOR_PATH) --config $(CONFIG_ROTATOR_PATH) run
docker-up: os
	docker-compose -f docker-compose.yml up -d
docker-down:
	docker-compose -f docker-compose.yml down -v
.PHONY: build, all, fmt, lint, test, run, os,  generate