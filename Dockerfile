FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o posd ./cmd/posd

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/posd .
COPY assets/ ./assets/
COPY migrations/ ./migrations/

EXPOSE 8080

CMD ["./posd"]
