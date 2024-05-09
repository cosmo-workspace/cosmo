.PHONY: lint test vendor clean

default: lint test

lint:
	golangci-lint run

test:
	go test -v -cover ./...

yaegi_test:
	yaegi test -v .

vendor:
	GOWORK=off go mod vendor

clean:
	rm -rf ./vendor