BINARY  := liftoff-replay
CMD     := ./cmd/$(BINARY)
BIN     := bin
DIST    := dist

.PHONY: build run test build-all clean

build:
	mkdir -p $(BIN)
	CGO_ENABLED=0 go build -o $(BIN)/$(BINARY) $(CMD)

run:
	go run $(CMD)

test:
	go test ./...

build-all:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o $(DIST)/$(BINARY)-linux-amd64        $(CMD)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -o $(DIST)/$(BINARY)-linux-arm64        $(CMD)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o $(DIST)/$(BINARY)-darwin-amd64       $(CMD)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o $(DIST)/$(BINARY)-darwin-arm64       $(CMD)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(DIST)/$(BINARY)-windows-amd64.exe  $(CMD)

clean:
	rm -rf $(BIN) $(DIST)

