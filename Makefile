.PHONY: all
all: local-run

APP_NAME=minimal-server-monitoring

.PHONY: $(APP_NAME) build test-coverage test lint generate local-run docker-run

build:
	go build -o $(APP_NAME) ./cmd/$(APP_NAME)

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test:
	go test ./...
	
lint: tools/bin/golangci-lint
	tools/bin/golangci-lint run ./...

tools/bin/golangci-lint: tools/go.mod tools/go.sum
	cd tools && go build -o bin/golangci-lint github.com/golangci/golangci-lint/v2/cmd/golangci-lint

generate:
	go generate ./...

local-run:
	go run -race cmd/$(APP_NAME)/$(APP_NAME).go config.yml

docker-run:
	docker build -t $(APP_NAME) .
	docker run -e SHOUTRRR=$(SHOUTRRR) -e MACHINENAME=$(shell hostname) -it -v /var/run/docker.sock:/var/run/docker.sock:ro -v /run/systemd:/run/systemd:ro -v /:/host:ro $(APP_NAME)
