# --- BUILD STAGE --- #
FROM golang:1.23.5-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o rss-grpc *.go

# --- RUNNER STAGE --- #
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/rss-grpc .

EXPOSE 50051
CMD ["./rss-grpc"]
