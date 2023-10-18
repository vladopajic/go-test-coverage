GO ?= go
GOBIN ?= $$($(GO) env GOPATH)/bin
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.54.2

.PHONY: get-golangcilint
get-golangcilint:
	test -f $(GOLANGCI_LINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$($(GO) env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

# Runs lint on entire repo
.PHONY: lint
lint: get-golangcilint
	$(GOLANGCI_LINT) run ./...

# Runs tests on entire repo
.PHONY: test
test: 
	go test -timeout=3s -race -count=10 -failfast -short ./...
	go test -timeout=3s -race -count=1 -failfast ./...

# Code tidy
.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

# Runs test coverage check
.PHONY: check-coverage
check-coverage:
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	go run ./main.go --config=./.github/.testcoverage.yml