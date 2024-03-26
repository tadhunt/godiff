export GOPRIVATE=github.com/tadhunt

all:
	go mod tidy
	go vet
	go build

test: all
	go test -v ./...
