FROM golang:1.22-bullseye AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o redisecho .

FROM debian:bullseye-slim
WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    netcat \
    telnet \
    iputils-ping && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/redisecho .

EXPOSE 9122
CMD ["./redisecho"]
