.PHONY: lint test vendor clean

default: lint test

lint:
	golangci-lint run

test:
	go test -v -cover ./...

yaegi_test:
	yaegi test -v .

vendor:
	go work vendor

clean:
	rm -rf ./vendor