BIN      := bin
DIST     := dist
CMDS     := $(notdir $(wildcard cmd/*))

GO       := go
CGO      := 0

.PHONY: build run test clean build-all

build:
	mkdir -p $(BIN)
	@for cmd in $(CMDS); do \
		echo "Building $$cmd"; \
		CGO_ENABLED=$(CGO) $(GO) build -o $(BIN)/$$cmd ./cmd/$$cmd; \
	done

run:
	$(GO) run ./cmd/liftoff-telemetry

test:
	$(GO) test ./...

build-all:
	mkdir -p $(DIST)
	@for cmd in $(CMDS); do \
		echo "Cross-building $$cmd"; \
		CGO_ENABLED=$(CGO) GOOS=linux   GOARCH=amd64 $(GO) build -o $(DIST)/$$cmd-linux-amd64        ./cmd/$$cmd; \
		CGO_ENABLED=$(CGO) GOOS=darwin  GOARCH=amd64 $(GO) build -o $(DIST)/$$cmd-darwin-amd64       ./cmd/$$cmd; \
		CGO_ENABLED=$(CGO) GOOS=windows GOARCH=amd64 $(GO) build -o $(DIST)/$$cmd-windows-amd64.exe  ./cmd/$$cmd; \
	done

clean:
	rm -rf $(BIN) $(DIST)
