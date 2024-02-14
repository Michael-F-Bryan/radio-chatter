# syntax=docker/dockerfile:1

FROM golang:1.21 AS build

RUN mkdir /app
WORKDIR /app

COPY go.* /app/
RUN go mod download -x
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download -x
COPY cmd/ /app/cmd/
COPY pkg/ /app/pkg/
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go build -v -o ./out/ ./...

######

FROM ubuntu:24.04

COPY --from=build /app/out/radio-chatter /bin/radio-chatter
EXPOSE 8080

# Note: we use curl for healthchecks
RUN apt-get update && apt-get install -y curl ffmpeg && rm -rf /var/lib/apt/lists/*
