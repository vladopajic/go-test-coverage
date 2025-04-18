package main

import (
	"fmt"
	"os"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/logger"
)

const (
	Version = "v2.13.2" // VERSION: when changing version update version in other places
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

	pass, haderr := testcoverage.Check(os.Stdout, cfg)
	if haderr {
		fmt.Println("\nRunning coverage check failed. Please use --debug=true flag to see detailed output.")
	}
	if !pass || haderr {
		os.Exit(1)
	}
}
