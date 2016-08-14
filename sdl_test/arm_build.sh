#!/bin/bash

#sudo apt-get install gcc-arm-linux-gnueabihf
#CXX=android-armeabi-g++ \
#GOARM=7

env CGO_ENABLED=1 \
GOOS=linux \
GOARCH=arm \
CC=arm-linux-gnueabihf-gcc \
go build -v

