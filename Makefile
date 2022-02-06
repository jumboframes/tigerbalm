all: build

build:
	go build -ldflags "-s -w" -o tigerbalm cmd/main.go

clean:
	rm tigerbalm

output: build
