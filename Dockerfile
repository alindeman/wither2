FROM golang:1.13 AS build

WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install -v ./cmd/...

FROM debian:buster-slim
RUN apt-get update \
  && apt-get install -y ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=build /go/bin/wither2 /usr/bin
