FROM golang:1.13 AS build

WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install -v ./cmd/...

FROM debian:buster-slim
COPY --from=build /go/bin/wither2 /usr/bin
