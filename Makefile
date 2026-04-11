.PHONY: run build seed reset tidy swagger test

run:
	go run ./cmd

build:
	go build -o bin/subscription_service ./cmd

seed:
	go run ./cmd seed

reset:
	go run ./cmd reset

tidy:
	go mod tidy

swagger:
	$(shell go env GOPATH)/bin/swag init -g cmd/main.go --output docs

test:
	go test ./internal/usecase/... -v -race
