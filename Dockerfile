# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.25.4-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/main ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=build /app/main .
COPY --from=build /app/web/templates ./web/templates

EXPOSE 8080

CMD ["./main"]
