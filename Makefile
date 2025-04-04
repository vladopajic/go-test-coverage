GO ?= go
GOBIN ?= $$($(GO) env GOPATH)/bin
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.0.1 # LINT_VERSION: update version in other places

# Code tidy
.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

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
	go test -timeout=20s -race -count=1 -failfast  -shuffle=on ./... -coverprofile=./cover.profile -covermode=atomic -coverpkg=./...

# Runs test coverage check
.PHONY: check-coverage
check-coverage: test
	go run ./main.go --config=./.github/.testcoverage-local.yml

# View coverage profile
.PHONY: view-coverage
view-coverage:
	go tool cover -html=cover.profile -o=cover.html
	xdg-open cover.html

# Recreates badges-integration-test branch
.PHONY: new-branch-badges-integration-test
new-branch-badges-integration-test:
	git branch -D badges-integration-test
	git push origin --delete badges-integration-test
	git checkout --orphan badges-integration-test
	git rm -rf .
	echo "# Badges from Integration Tests" > README.md
	git add README.md
	git commit -m "initial commit"
	git push origin badges-integration-test