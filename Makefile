# Makefile for json2xml-go

# Variables
BINARY_NAME := json2xml-go
VERSION := 1.0.0
GO := go
GOFMT := gofmt
GOVET := go vet

# Installation directories
PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin
MANDIR := $(PREFIX)/share/man/man1

# Build info
LDFLAGS := -ldflags "-s -w"

.PHONY: all build clean install uninstall fmt vet test man install-man help

all: build

## build: Build the binary
build:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/json2xml-go

## install: Install binary and man page
install: build install-man
	@echo "Installing $(BINARY_NAME) to $(BINDIR)..."
	@mkdir -p $(BINDIR)
	@install -m 755 $(BINARY_NAME) $(BINDIR)/$(BINARY_NAME)
	@echo "Installation complete!"
	@echo ""
	@echo "Binary installed to: $(BINDIR)/$(BINARY_NAME)"
	@echo "Man page installed to: $(MANDIR)/$(BINARY_NAME).1"
	@echo ""
	@echo "Try: $(BINARY_NAME) --help"
	@echo "     man $(BINARY_NAME)"

## install-man: Install man page only
install-man:
	@echo "Installing man page to $(MANDIR)..."
	@mkdir -p $(MANDIR)
	@install -m 644 man/$(BINARY_NAME).1 $(MANDIR)/$(BINARY_NAME).1
	@echo "Man page installed."

## uninstall: Remove binary and man page
uninstall:
	@echo "Removing $(BINARY_NAME)..."
	@rm -f $(BINDIR)/$(BINARY_NAME)
	@rm -f $(MANDIR)/$(BINARY_NAME).1
	@echo "Uninstall complete."

## clean: Remove build artifacts
clean:
	@rm -f $(BINARY_NAME)
	@$(GO) clean

## fmt: Format code
fmt:
	$(GOFMT) -w .

## vet: Run go vet
vet:
	$(GOVET) ./...

## test: Run tests
test:
	$(GO) test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## man: View man page locally (without installing)
man:
	@man ./man/$(BINARY_NAME).1

## help: Show this help
help:
	@echo "json2xml-go Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Variables:"
	@echo "  PREFIX    Installation prefix (default: /usr/local)"
	@echo "  BINDIR    Binary directory (default: \$$PREFIX/bin)"
	@echo "  MANDIR    Man page directory (default: \$$PREFIX/share/man/man1)"
	@echo ""
	@echo "Examples:"
	@echo "  make                    # Build the binary"
	@echo "  make install            # Install to /usr/local (may need sudo)"
	@echo "  sudo make install       # Install with root permissions"
	@echo "  make PREFIX=~/.local install  # Install to ~/.local"
	@echo "  make uninstall          # Remove installed files"
	@echo "  make man                # Preview man page"
