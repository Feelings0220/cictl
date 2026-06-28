BINARY := cictl
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test lint clean vendor-to-kagent

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/cictl

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY) $(BINARY).exe coverage.txt

# Stage the examples/jenkins-triage walkthrough into a kagent worktree at $(KAGENT_DIR).
# Used to produce the kagent PR. See docs/kagent-pr-checklist.md.
vendor-to-kagent:
	@test -n "$(KAGENT_DIR)" || (echo "set KAGENT_DIR=path/to/kagent"; exit 1)
	mkdir -p $(KAGENT_DIR)/examples/jenkins-triage
	cp -r kagent-pr/examples-jenkins-triage/* $(KAGENT_DIR)/examples/jenkins-triage/
