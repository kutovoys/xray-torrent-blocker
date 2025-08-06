.PHONY: all amd64 arm64 test test-coverage test-short clean

all: amd64 arm64

amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=v0.0.1dev" -o build/tblocker_amd64

arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.Version=v0.0.1dev" -o build/tblocker_arm64

test:
	go test ./... -v

test-short:
	go test ./... -v -short

clean:
	rm -rf build/
	rm -f coverage.out coverage.html