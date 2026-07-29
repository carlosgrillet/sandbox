BINDIR      := $(CURDIR)/bin
BINNAME     := sb

GOFLAGS     :=
TAGS        :=
LDFLAGS     := -w -s
CGO_ENABLED ?= 0

.PHONY: build
build: $(BINDIR)/$(BINNAME) tidy

$(BINDIR)/$(BINNAME): $(SRC)
	CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) -trimpath -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o '$(BINDIR)'/$(BINNAME) .

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	@rm -rf '$(BINDIR)'
