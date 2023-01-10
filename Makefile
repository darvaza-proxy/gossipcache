.PHONY: all clean generate fmt
.PHONY: tidy get build test up

GO ?= go
GOFMT ?= gofmt
GOFMT_FLAGS = -w -l -s
GOGENERATE_FLAGS = -v

GOBIN ?= $(GOPATH)/bin

REVIVE ?= $(GOBIN)/revive
REVIVE_FLAGS ?= -formatter friendly
REVIVE_INSTALL_URL ?= github.com/mgechev/revive

TMPDIR ?= .tmp

all: get generate tidy build

clean:
	rm -rf $(TMPDIR)

fmt:
	@find . -name '*.go' | xargs -r $(GOFMT) $(GOFMT_FLAGS)

tidy: fmt $(REVIVE)
	$(GO) mod tidy
	$(GO) vet ./...
	$(REVIVE) $(REVIVE_RUN_ARGS) ./...

get:
	$(GO) get -v ./...

build:
	$(GO) build -v ./...

test:
	$(GO) test -v ./...

up:
	$(GO) get -u -v ./...
	$(GO) mod tidy
	$(GO) install -v $(REVIVE_INSTALL_URL)

generate:
	$(GO) generate $(GOGENERATE_FLAGS) ./...

$(REVIVE):
	$(GO) install -v $(REVIVE_INSTALL_URL)
