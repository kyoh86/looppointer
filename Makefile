.PHONY: gen lint test install man

VERSION := `git vertag get`
COMMIT  := `git rev-parse HEAD`

gen:
	go generate ./...

lint: gen
	golangci-lint run

test: gen
	go test -v --race ./...

install: test
	go install -a -ldflags "-X=main.version=$(VERSION) -X=main.commit=$(COMMIT)" ./...

plugin: test
	go build -o looppointer.so -buildmode=plugin ./plugin
