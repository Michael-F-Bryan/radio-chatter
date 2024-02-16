# syntax=docker/dockerfile:1

FROM golang:1.21 AS build

RUN mkdir /app
WORKDIR /app

# Download and compile all dependencies
COPY go.* /app/
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download -x && echo "package dummy" > dummy.go && go build -v "./..." && rm dummy.go

# Build the main package
COPY cmd/ /app/cmd/
COPY pkg/ /app/pkg/
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go build -v -o ./out/ ./...

######

FROM python:3.12-slim AS transcribe

# Set up Whisper
RUN apt-get update && apt-get install -y git ffmpeg && rm -rf /var/lib/apt/lists/*
RUN pip3 install "git+https://github.com/openai/whisper.git"
RUN whisper --version
RUN python3 -c "import whisper; whisper.load_model('large-v2', device='cpu')"

COPY --from=build /app/out/radio-chatter /bin/radio-chatter

######

FROM ubuntu:24.04

COPY --from=build /app/out/radio-chatter /bin/radio-chatter
EXPOSE 8080

# Note: we use curl for healthchecks
RUN apt-get update && apt-get install -y curl ffmpeg && rm -rf /var/lib/apt/lists/*
