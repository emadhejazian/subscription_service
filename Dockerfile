FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o subscription_service ./cmd

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/subscription_service .
EXPOSE 8080
ENTRYPOINT ["./subscription_service"]
