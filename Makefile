.PHONY: run build seed tidy

run:
	go run ./cmd/main.go

build:
	go build -o bin/subscription_service ./cmd/main.go

seed:
	go run ./cmd/main.go --seed

tidy:
	go mod tidy
