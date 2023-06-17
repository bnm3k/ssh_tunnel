.SILENT:
.DEFAULT_GOAL:=build

.PHONY: build install help clean

## build: build ssh_tunnel
build:
	go build .

## install: install ssh_tunnel to $GOPATH/bin
install:
	go install .

## clean: remove ssh_tunnel
clean:
	rm ssh_tunnel

## help: print this help message
help:
	echo 'Usage:'
	sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
