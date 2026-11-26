FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o main cmd/api-server/main.go

FROM ubuntu:22.04 AS mysql-binlog

RUN apt-get update && apt-get install -y --no-install-recommends mysql-server-core-8.0 && \
    mkdir /binlog && cp /usr/bin/mysqlbinlog /binlog/ && \
    ldd /usr/bin/mysqlbinlog | awk '/=> \// { print $3 } /^[[:space:]]*\/.*ld-linux/ { print $1 }' | xargs -r -I '{}' cp -L '{}' /binlog/ && \
    mv /binlog/ld-linux-*.so.* /binlog/loader

FROM alpine:latest

RUN apk add --no-cache \
    sqlite-libs \
    postgresql-client \
    mysql-client \
    mongodb-tools \
    mariadb-connector-c

COPY --from=mysql-binlog /binlog /opt/mysql-binlog
RUN printf '%s\n' '#!/bin/sh' 'exec /opt/mysql-binlog/loader --library-path /opt/mysql-binlog /opt/mysql-binlog/mysqlbinlog "$@"' > /usr/local/bin/mysqlbinlog && \
    chmod +x /usr/local/bin/mysqlbinlog && mysqlbinlog --version

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/internal/database ./internal/database

EXPOSE 8080

CMD ["./main"]
