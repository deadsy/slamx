
all: pc

pc:
	go install -v -tags pc

rpi:
	go install -v -tags rpi

clean:
	go clean

.PHONY: all pc rpi clean
