FROM golang:alpine AS builder

RUN go version

RUN apk update && apk upgrade && apk add git zlib-dev gcc musl-dev

COPY . /go/src/github.com/jadevelopmentgrp/Tickets-Archiver
WORKDIR  /go/src/github.com/jadevelopmentgrp/Tickets-Archiver

RUN set -Eeux && \
    go mod download && \
    go mod verify

RUN GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -o Tickets-Archiver cmd/Tickets-Archiver/main.go

FROM alpine:latest

RUN apk update && apk upgrade

COPY --from=builder /go/src/jadevelopmentgrp/Tickets-Archiver/Tickets-Archiver /srv/Tickets-Archiver/Tickets-Archiver
RUN chmod +x /srv/Tickets-Archiver/Tickets-Archiver

RUN adduser container --disabled-password --no-create-home
USER container
WORKDIR /srv/Tickets-Archiver

CMD ["/srv/Tickets-Archiver/Tickets-Archiver"]
