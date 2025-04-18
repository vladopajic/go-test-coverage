package main

import (
	"fmt"
	"os"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/logger"
)

const (
	Version = "v2.14.0" // VERSION: when changing version update version in other places
	Name    = "go-test-coverage"
)

//nolint:forbidigo,wsl // relax
func main() {
	cfg, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	logger.Init()

	pass, err := testcoverage.Check(os.Stdout, cfg)
	if err != nil {
		fmt.Println("Running coverage check failed.")
		if cfg.GithubActionOutput {
			fmt.Printf("Please set `debug: true` input to see detailed output.")
		} else {
			fmt.Println("Please use `--debug=true` flag to see detailed output.")
		}
	}
	if !pass || err != nil {
		os.Exit(1)
	}
}
