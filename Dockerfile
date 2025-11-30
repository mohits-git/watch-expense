FROM golang:1.24.4 AS builder

WORKDIR /usr/src/watchexpense

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/web

FROM alpine:latest

RUN apk add --no-cache curl

WORKDIR /usr/bin

COPY --from=builder /usr/src/watchexpense/app /usr/bin/app

CMD [ "app" ]
