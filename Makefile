
all: dev

dev:
	env GOOS=linux GOARCH=amd64 go build -v -tags dev

rpi:
	env GOOS=linux GOARCH=arm go build -v -tags rpi

clean:
	go clean

.PHONY: all dev rpi clean
