FROM golang:alpine AS builder

RUN go version

RUN apk update && apk upgrade && apk add git zlib-dev gcc musl-dev

COPY . /go/src/logarchiver
WORKDIR  /go/src/logarchiver

RUN set -Eeux && \
    go mod download && \
    go mod verify

RUN GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -o logarchiver cmd/logarchiver/main.go

FROM alpine:latest

RUN apk update && apk upgrade

COPY --from=builder /go/src/logarchiver/logarchiver /srv/logarchiver/logarchiver
RUN chmod +x /srv/logarchiver/logarchiver

RUN adduser container --disabled-password --no-create-home
USER container
WORKDIR /srv/logarchiver

CMD ["/srv/logarchiver/logarchiver"]