all: test lint style

setup:
	go get -d -t -v ./...

test: setup
	go test .

style: test
	gofmt -w .

lint: test
	golint ./...

.PHONY: test lint style setup
