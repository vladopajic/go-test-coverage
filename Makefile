GO ?= go
GOBIN ?= $$($(GO) env GOPATH)/bin
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.56.0

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
	go test -timeout=3s -race -count=10 -failfast -shuffle=on -short ./...
	go test -timeout=10s -race -count=1 -failfast  -shuffle=on ./...

# Code tidy
.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

# Runs test coverage check
.PHONY: generate-coverage
generate-coverage:
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...

# Runs test coverage check
.PHONY: check-coverage
check-coverage: generate-coverage
	go run ./main.go --config=./.github/.testcoverage.yml

# View coverage profile
.PHONY: view-coverage
view-coverage: generate-coverage
	go tool cover -html=cover.out -o=cover.html
	xdg-open cover.html