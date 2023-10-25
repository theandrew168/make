.POSIX:
.SUFFIXES:

.PHONY: default
default: build

.PHONY: build
build:
	go build -o make make.go

.PHONY: test
test:
	go test -count=1 ./...

.PHONY: race
race:
	go test -race -count=1 ./...

.PHONY: cover
cover:
	go test -coverprofile=c.out -coverpkg=./... -count=1 ./...
	go tool cover -html=c.out

.PHONY: format
format:
	gofmt -l -s -w .

.PHONY: update
update:
	go get -u ./...
	go mod tidy

.PHONY: clean
clean:
	rm -fr make c.out
