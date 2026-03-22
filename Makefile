# ytdl — local dev and Linux (amd64) release builds.
#
# Override on the command line, e.g.:
#   make run YTD_LISTEN=:8080 YTD_DOWNLOAD_DIR=~/Downloads/ytdl

.PHONY: all build run linux-amd64 vet fmt test tidy clean help

CMD          := ./cmd/ytdl
BIN_DIR      := bin
BINARY       := ytdl
LINUX_BINARY := $(BINARY)-linux-amd64

YTD_DOWNLOAD_DIR ?= $(CURDIR)/tmp/downloads
YTD_LISTEN       ?= 127.0.0.1:8080

all: build

help:
	@echo "Targets:"
	@echo "  make build         Build $(BIN_DIR)/$(BINARY) for this machine (darwin/linux native)"
	@echo "  make run           go run with YTD_DOWNLOAD_DIR=$(YTD_DOWNLOAD_DIR)"
	@echo "  make linux-amd64   Cross-build $(BIN_DIR)/$(LINUX_BINARY) for Debian/x86_64 (Xeon, etc.)"
	@echo "  make vet fmt test tidy  — usual Go checks"
	@echo "  make clean         Remove $(BIN_DIR)/"
	@echo ""
	@echo "Variables: YTD_LISTEN YTD_DOWNLOAD_DIR (for run)"

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

build: $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(CMD)

linux-amd64: $(BIN_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BIN_DIR)/$(LINUX_BINARY) $(CMD)
	@echo "Built $(BIN_DIR)/$(LINUX_BINARY) — copy to Debian, e.g. /usr/local/bin/$(BINARY)"

run:
	mkdir -p $(YTD_DOWNLOAD_DIR)
	YTD_DOWNLOAD_DIR=$(YTD_DOWNLOAD_DIR) YTD_LISTEN=$(YTD_LISTEN) go run $(CMD)

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal web

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)
