FROM golang:1.25-alpine AS builder

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

FROM mongo:8.0 AS mongo-tools

RUN mkdir /mongo-tools && \
    for tool in mongodump mongorestore; do \
        bin="$(command -v "$tool")" && cp "$bin" /mongo-tools/ && \
        ldd "$bin" | awk '/=> \// { print $3 } /^[[:space:]]*\/.*ld-linux/ { print $1 }' | xargs -r -I '{}' cp -L '{}' /mongo-tools/ || exit 1; \
    done && mv /mongo-tools/ld-linux-*.so.* /mongo-tools/loader

FROM alpine:latest

RUN apk add --no-cache \
    sqlite-libs \
    postgresql18-client \
    mariadb-client \
    mariadb-connector-c

COPY --from=mongo-tools /mongo-tools /opt/mongo-tools
RUN for tool in mongodump mongorestore; do \
        printf '#!/bin/sh\nexec /opt/mongo-tools/loader --library-path /opt/mongo-tools /opt/mongo-tools/%s "$@"\n' "$tool" > "/usr/local/bin/$tool" && \
        chmod +x "/usr/local/bin/$tool" && "$tool" --version || exit 1; \
    done

COPY --from=mysql-binlog /binlog /opt/mysql-binlog
RUN printf '%s\n' '#!/bin/sh' 'exec /opt/mysql-binlog/loader --library-path /opt/mysql-binlog /opt/mysql-binlog/mysqlbinlog "$@"' > /usr/local/bin/mysqlbinlog && \
    chmod +x /usr/local/bin/mysqlbinlog && mysqlbinlog --version

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/internal/database ./internal/database

EXPOSE 8080

CMD ["./main"]
