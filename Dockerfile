FROM golang:1.22.3-alpine3.19

RUN apk update && apk add bash

WORKDIR /app

COPY ./api .
COPY ./api/go.mod ./api/go.sum ./
COPY ./api/.air.toml ./

RUN go mod download && \
    go build -o myapp ./cmd && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/cosmtrek/air@latest

# airを起動
CMD ["air", "-c", ".air.toml"]
