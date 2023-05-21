build:
	@echo go build -o "bin/go-image-server_$$(git describe --tags --always --dirty)" . 
	@go build -o "bin/go-image-server_$$(git describe --tags --always --dirty)" .

test:
	go test -v ./...

test-race:
	go test -race ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf bin
	rm -rf coverage.out
	rm -rf test-fs/*
