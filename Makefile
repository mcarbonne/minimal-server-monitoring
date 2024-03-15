.PHONY: all
all: local-run

APP_NAME=minimal-server-monitoring

.PHONY: $(APP_NAME)

build:
	go build -o $(APP_NAME) ./cmd/$(APP_NAME)

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test:
	go test ./...

local-run:
	go run cmd/$(APP_NAME)/$(APP_NAME).go config.json

docker-run:
	docker build -t $(APP_NAME) .
	docker run -e SHOUTRRR=$(SHOUTRRR) -it -v /var/run/docker.sock:/var/run/docker.sock:ro $(APP_NAME)
