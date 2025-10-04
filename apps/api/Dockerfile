FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o main cmd/api-server/main.go

FROM alpine:latest

RUN apk add --no-cache \
    sqlite-libs \
    postgresql-client \
    mysql-client \
    mongodb-tools

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/internal/database ./internal/database

EXPOSE 8080

CMD ["./main"]
