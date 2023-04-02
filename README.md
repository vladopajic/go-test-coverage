# go-test-coverage

[![test](https://github.com/vladopajic/go-test-coverage/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-test-coverage/actions/workflows/test.yml)
[![lint](https://github.com/vladopajic/go-test-coverage/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-test-coverage/actions/workflows/lint.yml)
 [![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/main/coverage.svg)](https://github.com/vladopajic/go-test-coverage/tree/badges)
[![Go Report Card](https://goreportcard.com/badge/github.com/vladopajic/go-test-coverage?cache=v1)](https://goreportcard.com/report/github.com/vladopajic/go-test-coverage)
[![GoDoc](https://godoc.org/github.com/vladopajic/go-test-coverage?status.svg)](https://godoc.org/github.com/vladopajic/go-test-coverage)
[![Release](https://img.shields.io/github/release/vladopajic/go-test-coverage.svg?style=flat-square)](https://github.com/vladopajic/go-test-coverage/releases/latest)


`go-test-coverage` is tool which reports issues when test coverage of a file or package is below set threshold.

### Usage

```yml
name: Go test coverage check
runs-on: ubuntu-latest
steps:
  - uses: actions/checkout@v3
  - uses: actions/setup-go@v3
  
  - name: test (generate coverage)
    run: go test ./... -coverprofile=./cover.out

  - name: check test coverage
    uses: vladopajic/go-test-coverage@v2
    with:
      # Configure with config file (option 1, has priority over option 2)
      config: ./.testcoverage.yml
      
      # Specify each config value (option 2)
      # `config` input has to be empty in order for these inputs to be used
      profile: cover.out
      local-prefix: github.com/org/project
      threshold-file: 80
      threshold-package: 80
      threshold-total: 95
```

### Config
Example of [.testcoverage.yml](./.testcoverage.example.yml) config file.

```yml
# (mandatory) 
# Path to coverprofile file (output of `go test -coverprofile` command)
profile: cover.out

# (optional; default false)
# When set to `true` tool will output github-action friendly outputs
github-action-output: true

# (optional) 
# When specified reported file paths will not contain local prefix in the output
local-prefix: "github.com/org/project"

# Holds coverage thresholds percentages, values should be in range [0-100]
threshold:
  # (optional; default 0) 
  # The minimum coverage that each file should have
  file: 80

  # (optional; default 0) 
  # The minimum coverage that each package should have
  package: 80

  # (optional; default 0) 
  # The minimum total coverage project should have
  total: 95
```

## Contribution

All contributions are useful, whether it is a simple typo, a more complex change, or just pointing out an issue. We welcome any contribution so feel free to open PR or issue. 
