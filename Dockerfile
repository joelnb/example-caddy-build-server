FROM golang:alpine3.20 AS builder

ENV CGO_ENABLED=0
RUN go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest

EXPOSE 8081

WORKDIR /app
COPY . /app
RUN go build -o caddy-build-server .

CMD ["./caddy-build-server"]
