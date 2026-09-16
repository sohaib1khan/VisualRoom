.PHONY: build test install uninstall dist clean

PREFIX ?= $(HOME)/.local/bin
GOFLAGS ?= -trimpath
LDFLAGS ?= -s -w
APP := vroom

build:
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(APP) .

test:
	go test ./...

install: build
	mkdir -p "$(PREFIX)"
	install -m 755 $(APP) "$(PREFIX)/$(APP)"
	@echo "installed $(PREFIX)/$(APP)"

uninstall:
	rm -f "$(PREFIX)/$(APP)"

dist:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/$(APP)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/$(APP)-linux-arm64 .

clean:
	rm -f $(APP)
	rm -rf dist
