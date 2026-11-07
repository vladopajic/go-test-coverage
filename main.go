package main

import (
	"fmt"
	"os"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/logger"
)

const Name = "go-test-coverage"

// TOOL_VERSION: when changing version update version in other places
// Version is overridden at image build time with -ldflags.
//
//nolint:gochecknoglobals // realx
var Version = "v2.22.0"

//nolint:forbidigo,wsl // relax
func main() {
	cfg, err := readConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	logger.Init()

	pass, err := testcoverage.Check(os.Stdout, cfg)
	if err != nil {
		fmt.Println("Running coverage check failed.")
		fmt.Printf("Error: %v\n", err)

		if cfg.GithubActionOutput {
			fmt.Println("Please set `debug: true` input to see detailed output.")
		} else {
			fmt.Println("Please use `--debug=true` flag to see detailed output.")
		}
	}

	if !pass || err != nil {
		os.Exit(1)
	}
}
