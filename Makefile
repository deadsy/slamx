
OUT = slamx
URL = 192.168.1.7

all: pc

pc:
	env GOOS=linux GOARCH=amd64 go build -v -tags pc

rpi:
	env GOOS=linux GOARCH=arm go build -v -tags rpi

copy:
	scp $(OUT) jasonh@$(URL):/home/jasonh/work/slamx

clean:
	go clean

.PHONY: all pc rpi clean
