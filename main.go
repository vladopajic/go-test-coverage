package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alexflint/go-arg"

	"github.com/vladopajic/go-test-coverage/pkg/testcoverage"
)

// Version value is injected at build time
//
//nolint:gochecknoglobals // must be global var
var Version string

//nolint:gochecknoinits // relax
func init() {
	if Version == "" {
		Version = "unknown-" + strconv.Itoa(int(time.Now().Unix()))
	}
}

type args struct {
	ConfigPath         string `arg:"-c,--config"`
	Profile            string `arg:"-p,--profile" help:"path to coverage profile"`
	LocalPrefix        string `arg:"-l,--local-prefix"`
	GithubActionOutput bool   `arg:"-o,--github-action-output"`
	ThresholdFile      int    `arg:"-f,--threshold-file"`
	ThresholdPackage   int    `arg:"-k,--threshold-package"`
	ThresholdTotal     int    `arg:"-t,--threshold-total"`
}

func (args) Version() string {
	return "go-test-coverage " + Version
}

func (a *args) toConfig() testcoverage.Config {
	cfg := testcoverage.NewConfig()

	cfg.Profile = fromMagicToEmpty(a.Profile)
	cfg.GithubActionOutput = a.GithubActionOutput
	cfg.LocalPrefix = fromMagicToEmpty(a.LocalPrefix)
	cfg.Threshold.File = a.ThresholdFile
	cfg.Threshold.Package = a.ThresholdPackage
	cfg.Threshold.Total = a.ThresholdTotal

	return cfg
}

//nolint:forbidigo // relax
func main() {
	cfg, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	stats, err := testcoverage.GenerateCoverageStats(cfg.Profile)
	if err != nil {
		fmt.Printf("failed to generate coverage statistics: %v\n", err)
		os.Exit(1)
	}

	result := testcoverage.Analyze(cfg, stats)

	testcoverage.ReportForHuman(result, cfg)

	if cfg.GithubActionOutput {
		testcoverage.ReportForGithubAction(result, cfg)

		err := testcoverage.SetGithubActionOutput(result)
		if err != nil {
			fmt.Printf("failed setting github action output: %v\n", err)
			os.Exit(1)
		}
	}

	if !result.Pass() {
		os.Exit(1)
	}
}

func readConfig() (testcoverage.Config, error) {
	cfg := testcoverage.NewConfig()
	cmdArgs := args{
		GithubActionOutput: cfg.GithubActionOutput,
		ThresholdFile:      cfg.Threshold.File,
		ThresholdPackage:   cfg.Threshold.Package,
		ThresholdTotal:     cfg.Threshold.Total,
	}
	arg.MustParse(&cmdArgs)

	cfgPath := fromMagicToEmpty(cmdArgs.ConfigPath)
	if cfgPath != "" {
		err := testcoverage.ConfigFromFile(&cfg, cfgPath)
		if err != nil {
			return testcoverage.Config{}, fmt.Errorf("failed loading config from file: %w", err)
		}
	} else {
		cfg = cmdArgs.toConfig()
	}

	if err := cfg.Validate(); err != nil {
		return testcoverage.Config{}, fmt.Errorf("config file is not valid: %w", err)
	}

	return cfg, nil
}

func fromMagicToEmpty(s string) string {
	if s == `''` {
		return ""
	}

	return s
}
