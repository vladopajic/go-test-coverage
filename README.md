# go-test-coverage

[![test](https://github.com/vladopajic/go-test-coverage/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-test-coverage/actions/workflows/test.yml)
[![action-test](https://github.com/vladopajic/go-test-coverage/actions/workflows/action-test.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-test-coverage/actions/workflows/action-test.yml)
[![lint](https://github.com/vladopajic/go-test-coverage/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-test-coverage/actions/workflows/lint.yml)
[![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/main/coverage.svg)](/.github/.testcoverage.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vladopajic/go-test-coverage?cache=v1)](https://goreportcard.com/report/github.com/vladopajic/go-test-coverage)
[![Release](https://img.shields.io/github/release/vladopajic/go-test-coverage.svg?style=flat-square)](https://github.com/vladopajic/go-test-coverage/releases/latest)


`go-test-coverage` is tool which reports issues when test coverage is below set threshold.

## Why?

These are the most important features and benefits of `go-test-coverage`:

- quick, 5-minute installation
- server-less with no registration or permissions required
  -  check never fails due to connectivity/server issues
- **ensures data privacy**, no leaks to third parties
  - [risks of information leakage through remote code coverage services](https://gist.github.com/vladopajic/0b835b28bcfe4a5a22bb0ae20e365266)
- runs blazingly fast - (~1 sec on [go-test-coverage repo](https://github.com/vladopajic/go-test-coverage/actions/runs/8401578681/job/23010110385))
- versatile for local and CI use
- extensive configuration options
- stylish badges
- free and open-source!

## Usage

`go-test-coverage` can be used in two ways:
 - as local tool, and/or
 - as step of GitHub workflow

It is recommended to have both options in go repositories.

### Local tool

Example of `Makefile` which has `check-coverage` command that runs `go-test-coverage` locally:

```makefile
GOBIN ?= $$(go env GOPATH)/bin

.PHONY: install-go-test-coverage
install-go-test-coverage:
	go install github.com/vladopajic/go-test-coverage/v2@latest

.PHONY: check-coverage
check-coverage: install-go-test-coverage
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	${GOBIN}/go-test-coverage --config=./.testcoverage.yml
```

### Github workflow

Example to run `go-test-coverage` as step of workflow:


```yml
name: Go test coverage check
runs-on: ubuntu-latest
steps:
  - uses: actions/checkout@v3
  - uses: actions/setup-go@v3
  
  - name: generate test coverage
    run: go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...

  - name: check test coverage
    uses: vladopajic/go-test-coverage@v2
    with:
      # Configure action using config file (option 1)
      config: ./.testcoverage.yml
      
      # Configure action by specifying input parameters individually (option 2).
      # If you are using config file (option 1) you shouldn't use these parameters, however
      # specifting these action parameters will override appropriate config values.
      profile: cover.out
      local-prefix: github.com/org/project
      threshold-file: 80
      threshold-package: 80
      threshold-total: 95

      # Optionally set source directory
      source-dir: ./project
```

### Config

Example of [.testcoverage.yml](./.testcoverage.example.yml) config file:

```yml
# (mandatory) 
# Path to coverprofile file (output of `go test -coverprofile` command).
#
# For cases where there are many coverage profiles, such as when running 
# unit tests and integration tests separately, you can combine all those
# profiles into one. In this case, the profile should have a comma-separated list 
# of profile files, e.g., 'cover_unit.out,cover_integration.out'.
profile: cover.out

# (optional; but recommended to set) 
# When specified reported file paths will not contain local prefix in the output
local-prefix: "github.com/org/project"

# (optional) 
# When specified coverage check will use source code on specified path.
# This is only usefull for monorepo project.
source-dir: "./project"

# Holds coverage thresholds percentages, values should be in range [0-100]
threshold:
  # (optional; default 0) 
  # The minimum coverage that each file should have
  file: 70

  # (optional; default 0) 
  # The minimum coverage that each package should have
  package: 80

  # (optional; default 0) 
  # The minimum total coverage project should have
  total: 95

# Holds regexp rules which will override thresholds for matched files or packages 
# using their paths.
#
# First rule from this list that matches file or package is going to apply 
# new threshold to it. If project has multiple rules that match same path, 
# override rules should be listed in order from specific to more general rules.
override:
  # Increase coverage threshold to 100% for `foo` package 
  # (default is 80, as configured above in this example)
  - threshold: 100
    path: ^pkg/lib/foo$

# Holds regexp rules which will exclude matched files or packages 
# from coverage statistics
exclude:
  # Exclude files or packages matching their paths
  paths:
    - \.pb\.go$    # excludes all protobuf generated files
    - ^pkg/bar     # exclude package `pkg/bar`
 
# NOTES:
# - symbol `/` in all path regexps will be replaced by current OS file path separator
#   to properly work on Windows
```

### Exclude code from coverage statistics with comment annotation

For cases where there is a code block that does not need to be tested, it can be ignored from coverage statistics by adding the comment `// coverage-ignore` at the start line of the statement body (right after `{`).

```go
...
result, err := foo()
if err != nil { // coverage-ignore
	return err
}
...
```

Similarly, the entire function can be excluded from coverage statistics when a comment is found at the start line of the function body (right after `{`).
```go
func bar() { // coverage-ignore
...
}
```

## Coverage badge

Repositories which use `go-test-coverage` action in their workflows could easily create beautiful coverage badge and embed them in markdown files (eg. ![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/main/coverage.svg)).

Read instructions on creating coverage badge [here](./docs/badge.md).

## Visualise coverage profile

Go's toolchain includes a utility for visualizing coverage profiles, providing valuable insights into which statements have not been covered by tests. This feature proves highly beneficial for understanding the extent of test coverage in your codebase.

Following command will generate `cover.html` page with visualized coverage profile: 
```console
go tool cover -html=cover.out -o=cover.html
```


## Contribution

All contributions are useful, whether it is a simple typo, a more complex change, or just pointing out an issue. We welcome any contribution so feel free to open PR or issue. 
