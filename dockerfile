# TEST IMAGE
FROM golang:1.21.0 AS build
WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# test and build
COPY . .
RUN go test --race ./...
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o /usr/src/app/app .


# DEPLOYMENT IMAGE
FROM scratch AS deploy


COPY --from=build  /usr/src/app/app /
COPY imageServer_config.yaml imageServer_config.yaml
COPY docs /docs
COPY www /www

CMD ["/app"]