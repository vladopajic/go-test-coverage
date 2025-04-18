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

//nolint:forbidigo // relax
func main() {
	cfg, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	logger.Init()

	pass := testcoverage.Check(os.Stdout, cfg)
	if !pass {
		os.Exit(1)
	}
}
