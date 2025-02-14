.PHONY: all amd64 arm64

all: amd64 arm64

amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=v0.0.1dev" -o build/tblocker_amd64

arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.Version=v0.0.1dev" -o build/tblocker_arm64