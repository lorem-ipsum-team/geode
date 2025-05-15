BUILD_DIR=./bin
BIN_NAME=geode
BUILD_ENV=GOOS=linux GOARCH=amd64

.PHONY: lint format build download docker
lint:
	golangci-lint run

format:
	golangci-lint fmt
	golangci-lint run --fix

download:
	go mod download

build:
	$(BUILD_ENV) go build -o $(BUILD_DIR)/$(BIN_NAME) ./cmd/service/main.go

docker:
	docker build .