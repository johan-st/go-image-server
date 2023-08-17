# BUILD AND TEST IMAGE
FROM golang:1.21.0-alpine3.18 AS build
WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app .

CMD ["app"]


# DEPLOYMENT IMAGE

FROM alpine:3.18.3 AS deploy

WORKDIR /usr/src/app
COPY --from=build /usr/local/bin/app /usr/bin/app
COPY imageServer_config.yaml imageServer_config.yaml
COPY docs docs
COPY www www

CMD ["app"]