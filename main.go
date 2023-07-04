package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
)

const Version = "v2.7.1"

type args struct {
	ConfigPath         string `arg:"-c,--config"`
	Profile            string `arg:"-p,--profile" help:"path to coverage profile"`
	LocalPrefix        string `arg:"-l,--local-prefix"`
	GithubActionOutput bool   `arg:"-o,--github-action-output"`
	ThresholdFile      int    `arg:"-f,--threshold-file"`
	ThresholdPackage   int    `arg:"-k,--threshold-package"`
	ThresholdTotal     int    `arg:"-t,--threshold-total"`
}

func newArgs() args {
	return args{
		ConfigPath:         `''`,
		Profile:            `''`,
		LocalPrefix:        `''`,
		GithubActionOutput: false,
		ThresholdFile:      -1,
		ThresholdPackage:   -1,
		ThresholdTotal:     -1,
	}
}

func (args) Version() string {
	return "go-test-coverage " + Version
}

func (a *args) overrideConfig(cfg testcoverage.Config) testcoverage.Config {
	if !isMagicString(a.Profile) {
		cfg.Profile = a.Profile
	}

	if a.GithubActionOutput {
		cfg.GithubActionOutput = true
	}

	if !isMagicString(a.LocalPrefix) {
		cfg.LocalPrefix = a.LocalPrefix
	}

	if !isMagicInt(a.ThresholdFile) {
		cfg.Threshold.File = a.ThresholdFile
	}

	if !isMagicInt(a.ThresholdPackage) {
		cfg.Threshold.Package = a.ThresholdPackage
	}

	if !isMagicInt(a.ThresholdPackage) {
		cfg.Threshold.Total = a.ThresholdTotal
	}

	return cfg
}

//nolint:forbidigo // relax
func main() {
	cfg, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	result, err := testcoverage.Check(os.Stdout, cfg)
	if err != nil || !result.Pass() {
		if err != nil {
			fmt.Println(err.Error())
		}

		os.Exit(1)
	}
}

func readConfig() (testcoverage.Config, error) {
	cmdArgs := newArgs()
	arg.MustParse(&cmdArgs)

	cfg := testcoverage.Config{}

	// Load config from file
	if !isMagicString(cmdArgs.ConfigPath) {
		err := testcoverage.ConfigFromFile(&cfg, cmdArgs.ConfigPath)
		if err != nil {
			return testcoverage.Config{}, fmt.Errorf("failed loading config from file: %w", err)
		}
	}

	// Override config with values from args
	cfg = cmdArgs.overrideConfig(cfg)

	if err := cfg.Validate(); err != nil {
		return testcoverage.Config{}, fmt.Errorf("config file is not valid: %w", err)
	}

	return cfg, nil
}

func isMagicString(v string) bool {
	return v == `''`
}

func isMagicInt(v int) bool {
	return v == -1
}
