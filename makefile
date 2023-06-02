build: race
	@echo go build -o "bin/go-image-server_$$(git describe --tags --always --dirty)" . 
	@go build -o "bin/go-image-server_$$(git describe --tags --always --dirty)" .

run:
	go run . -c devConf.yaml 
	
dev:
	go run . -c devConf.yaml -dev

format:
	gofmt -s -w .

test:
	go test -vet=all ./...

race:
	go test -race ./...

bench:. 
	go test -benchmem -bench .

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf bin
	rm -rf coverage.out
	rm -rf test-fs/*
	rm *access.*