# BUILDER
FROM golang:1.21-bookworm as builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app .


# RUNNER
FROM debian:stable-slim as runner

WORKDIR /usr/src/app

# EMBED THESE FILES IN THE BINARY
COPY pages/assets pages/assets
COPY docs docs
COPY test-data test-data


COPY --from=builder /usr/local/bin/app /usr/local/bin/app


COPY prod.yaml config.yaml

CMD ["app", "-c", "config.yaml"]