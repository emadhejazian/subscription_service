.PHONY: run build seed tidy swagger

run:
	go run ./cmd

build:
	go build -o bin/subscription_service ./cmd

seed:
	go run ./cmd seed

tidy:
	go mod tidy

swagger:
	$(shell go env GOPATH)/bin/swag init -g cmd/main.go --output docs
