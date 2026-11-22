GOBIN ?= $$(go env GOPATH)/bin
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint
# LINT_VERSION: update version in other places
GOLANGCI_LINT_VERSION ?= v2.14.0

# Code tidy
.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

.PHONY: get-golangcilint
get-golangcilint:
	test -x "$(GOLANGCI_LINT)" || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh | sh -s -- -b "$(GOBIN)" $(GOLANGCI_LINT_VERSION)
	test -x "$(GOLANGCI_LINT)"

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
	go run ./ --config=./.github/.testcoverage-local.yml

# View coverage profile
.PHONY: view-coverage
view-coverage:
	test -f cover.profile || { echo "cover.profile missing - run 'make test' first"; exit 1; }
	go tool cover -html=cover.profile -o=cover.html
	xdg-open cover.html

# Recreates badges-integration-test branch
.PHONY: new-branch-badges-integration-test
new-branch-badges-integration-test:
	@set -e; \
		original_ref=$$(git symbolic-ref --quiet --short HEAD || git rev-parse --verify HEAD); \
		test "$$original_ref" != "badges-integration-test" || { echo "cannot recreate the currently checked-out badges-integration-test branch"; exit 1; }; \
		test -z "$$(git status --porcelain)" || { echo "working tree is dirty - commit, stash, or remove untracked files first"; exit 1; }; \
		trap 'git checkout -f "$$original_ref"' EXIT; \
		git branch -D badges-integration-test 2>/dev/null || true; \
		git checkout --orphan badges-integration-test; \
		git rm -rf .; \
		echo "# Badges from Integration Tests" > README.md; \
		git add README.md; \
		git commit -m "initial commit"; \
		git push --force origin badges-integration-test
