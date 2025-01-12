# go-test-coverage

[![test](https://github.com/vladopajic/go-test-coverage/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-test-coverage/actions/workflows/test.yml)
[![action-test](https://github.com/vladopajic/go-test-coverage/actions/workflows/action-test.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-test-coverage/actions/workflows/action-test.yml)
[![lint](https://github.com/vladopajic/go-test-coverage/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-test-coverage/actions/workflows/lint.yml)
[![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/main/coverage.svg)](/.github/.testcoverage.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vladopajic/go-test-coverage?cache=v1)](https://goreportcard.com/report/github.com/vladopajic/go-test-coverage)
[![Release](https://img.shields.io/github/release/vladopajic/go-test-coverage.svg?color=%23007ec6)](https://github.com/vladopajic/go-test-coverage/releases/latest)

![go-test-coverage cover image](https://github.com/user-attachments/assets/2febc74e-7437-4dc6-87a4-0ca47f8e714e)

`go-test-coverage` is a tool designed to report issues when test coverage falls below a specified threshold, ensuring higher code quality and preventing regressions in test coverage over time.

## Why Use go-test-coverage?

Here are the key features and benefits:

- **Quick Setup**: Install and configure in just 5 minutes.
- **Serverless Operation**: No need for external servers, registration, or permissions.
  - Eliminates connectivity or server-related failures.
- **Data Privacy**: All coverage checks are done locally, so no sensitive information leaks to third parties.
  - Learn more about [information leakage risks](https://gist.github.com/vladopajic/0b835b28bcfe4a5a22bb0ae20e365266).
- **Performance**: Lightning-fast execution (e.g., ~1 second on [this repo](https://github.com/vladopajic/go-test-coverage/actions/runs/8401578681/job/23010110385)).
- **Versatility**: Can be used both locally and in CI pipelines.
- **Customizable**: Extensive configuration options to fit any project's needs.
- **Stylish Badges**: Generate beautiful coverage badges for your repository.
- **Open Source**: Free to use and contribute to!

## Usage

You can use  `go-test-coverage` in two ways:
 - Locally as part of your development process.
 - As a step in your GitHub Workflow.

It’s recommended to utilize both options for Go projects.

### Local Usage

Here’s an example `Makefile` with a `check-coverage` command that runs `go-test-coverage` locally:


```makefile
GOBIN ?= $$(go env GOPATH)/bin

.PHONY: install-go-test-coverage
install-go-test-coverage:
	go install github.com/subhambhardwaj/go-test-coverage/v2@latest

.PHONY: check-coverage
check-coverage: install-go-test-coverage
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	${GOBIN}/go-test-coverage --config=./.testcoverage.yml
```

### GitHub Workflow

Here’s an example of how to integrate `go-test-coverage` into a GitHub Actions workflow:


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
      config: ./.testcoverage.yml
```

For detailed information about the GitHub Action, check out [this page](./docs/github_action.md).

### Configuration

Here’s an example [.testcoverage.yml](./.testcoverage.example.yml) configuration file:

```yml
# (mandatory) 
# Path to coverage profile file (output of `go test -coverprofile` command).
#
# For cases where there are many coverage profiles, such as when running 
# unit tests and integration tests separately, you can combine all those
# profiles into one. In this case, the profile should have a comma-separated list 
# of profile files, e.g., 'cover_unit.out,cover_integration.out'.
profile: cover.out

# (optional; but recommended to set) 
# When specified reported file paths will not contain local prefix in the output.
local-prefix: "github.com/org/project"

# Holds coverage thresholds percentages, values should be in range [0-100].
threshold:
  # (optional; default 0) 
  # Minimum coverage percentage required for individual files.
  file: 70

  # (optional; default 0) 
  # Minimum coverage percentage required for each package.
  package: 80

  # (optional; default 0) 
  # Minimum overall project coverage percentage required.
  total: 95

# Holds regexp rules which will override thresholds for matched files or packages 
# using their paths.
#
# First rule from this list that matches file or package is going to apply 
# new threshold to it. If project has multiple rules that match same path, 
# override rules should be listed in order from specific to more general rules.
override:
  # Increase coverage threshold to 100% for `foo` package 
  # (default is 80, as configured above in this example).
  - path: ^pkg/lib/foo$
    threshold: 100

# Holds regexp rules which will exclude matched files or packages 
# from coverage statistics.
exclude:
  # Exclude files or packages matching their paths
  paths:
    - \.pb\.go$    # excludes all protobuf generated files
    - ^pkg/bar     # exclude package `pkg/bar`

# File name of go-test-coverage breakdown file, which can be used to 
# analyze coverage difference.
breakdown-file-name: ''

diff:
  # File name of go-test-coverage breakdown file which will be used to 
  # report coverage difference.
  base-breakdown-file-name: ''
```

### Exclude Code from Coverage

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

## Generate Coverage Badge

You can easily generate a stylish coverage badge for your repository and embed it in your markdown files. Here’s an example badge: ![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/main/coverage.svg)

Instructions for badge creation are available [here](./docs/badge.md).

## Visualise Coverage

Go includes a built-in tool for visualizing coverage profiles, allowing you to see which parts of the code are not covered by tests. To generate a visual report:

Following command will generate `cover.html` page with visualized coverage profile: 
```console
go tool cover -html=cover.out -o=cover.html
```

## Support the Project

`go-test-coverage` is freely available for all users. If your organization benefits from this tool, especially if you’ve transitioned from a paid coverage service, consider [sponsoring the project](https://github.com/sponsors/vladopajic). 
Your sponsorship will help sustain development, introduce new features, and maintain high-quality support. Every contribution directly impacts the future growth and stability of this project.

## Contribution

We welcome all contributions - whether it's fixing a typo, adding new features, or pointing out an issue. Feel free to open a pull request or issue to contribute!


Happy coding 🌞
