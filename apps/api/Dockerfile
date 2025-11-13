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
    mongodb-tools \
    redis \
    freetds

# Install Microsoft SQL Server Tools
RUN apk add --no-cache curl gnupg && \
    curl -O https://download.microsoft.com/download/b/9/f/b9f3cce4-3925-46d4-9f46-da08869c6486/msodbcsql18_18.1.1.1-1_amd64.apk && \
    curl -O https://download.microsoft.com/download/b/9/f/b9f3cce4-3925-46d4-9f46-da08869c6486/mssql-tools18_18.1.1.1-1_amd64.apk && \
    apk add --allow-untrusted msodbcsql18_18.1.1.1-1_amd64.apk && \
    apk add --allow-untrusted mssql-tools18_18.1.1.1-1_amd64.apk && \
    rm -f msodbcsql18_18.1.1.1-1_amd64.apk mssql-tools18_18.1.1.1-1_amd64.apk && \
    apk del curl gnupg

ENV PATH="${PATH}:/opt/mssql-tools18/bin"

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/internal/database ./internal/database

EXPOSE 8080

CMD ["./main"]
