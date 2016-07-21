.PHONY: all deps test validate vet lint fmt

all: deps validate test-unit ## get dependencies, validate all checks and run tests

deps: ## get dependencies
	go get -t ./...
	go get github.com/golang/lint/golint

test-unit: ## run tests
	go test -timeout 10s -v -race -cover ./...

validate: vet lint fmt ## validate gofmt, golint and go vet

vet:
	go vet ./...

lint:
	out="$$(golint ./...)"; \
	if [ -n "$$(golint ./...)" ]; then \
		echo "$$out"; \
		exit 1; \
	fi

fmt:
	test -z "$(gofmt -s -l . | tee /dev/stderr)"

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
