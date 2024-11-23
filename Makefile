.PHONY: build run test clean

build:
	go build -o logo-generator main.go

test:
	go test ./...

clean:
	rm -f logo-generator
	rm -rf output cache