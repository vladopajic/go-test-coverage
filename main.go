package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
)

const (
	Version = "v2.8.0"
	Name    = "go-test-coverage"
)

const (
	// default value of string variables passed by CI
	ciDefaultString = `''`
	// default value of int variables passed by CI
	ciDefaultnt = -1
)

type args struct {
	ConfigPath         string `arg:"-c,--config"`
	Profile            string `arg:"-p,--profile" help:"path to coverage profile"`
	LocalPrefix        string `arg:"-l,--local-prefix"`
	GithubActionOutput bool   `arg:"-o,--github-action-output"`
	ThresholdFile      int    `arg:"-f,--threshold-file"`
	ThresholdPackage   int    `arg:"-k,--threshold-package"`
	ThresholdTotal     int    `arg:"-t,--threshold-total"`
	BadgeFileName      string `arg:"-b,--badge-file-name"`

	CDNKey            string `arg:"--cdn-key"`
	CDNSecret         string `arg:"--cdn-secret"`
	CDNRegion         string `arg:"--cdn-region"`
	CDNEndpoint       string `arg:"--cdn-endpoint"`
	CDNFileName       string `arg:"--cdn-file-name"`
	CDNBucketName     string `arg:"--cdn-bucket-name"`
	CDNForcePathStyle bool   `arg:"--cdn-force-path-style"`

	GitToken      string `arg:"--git-token"`
	GitRepository string `arg:"--git-repository"`
	GitBranch     string `arg:"--git-branch"`
	GitFileName   string `arg:"--git-file-name"`
}

func newArgs() args {
	return args{
		ConfigPath:         ciDefaultString,
		Profile:            ciDefaultString,
		LocalPrefix:        ciDefaultString,
		GithubActionOutput: false,
		ThresholdFile:      ciDefaultnt,
		ThresholdPackage:   ciDefaultnt,
		ThresholdTotal:     ciDefaultnt,

		// Badge
		BadgeFileName: ciDefaultString,

		// CDN
		CDNKey:            ciDefaultString,
		CDNSecret:         ciDefaultString,
		CDNRegion:         ciDefaultString,
		CDNEndpoint:       ciDefaultString,
		CDNFileName:       ciDefaultString,
		CDNBucketName:     ciDefaultString,
		CDNForcePathStyle: false,

		// Git
		GitToken:      ciDefaultString,
		GitRepository: ciDefaultString,
		GitBranch:     ciDefaultString,
		GitFileName:   ciDefaultString,
	}
}

func (args) Version() string {
	return Name + " " + Version
}

//nolint:cyclop,maintidx // relax
func (a *args) overrideConfig(cfg testcoverage.Config) testcoverage.Config {
	if !isCIDefaultString(a.Profile) {
		cfg.Profile = a.Profile
	}

	if a.GithubActionOutput {
		cfg.GithubActionOutput = true
	}

	if !isCIDefaultString(a.LocalPrefix) {
		cfg.LocalPrefix = a.LocalPrefix
	}

	if !isCIDefaultnt(a.ThresholdFile) {
		cfg.Threshold.File = a.ThresholdFile
	}

	if !isCIDefaultnt(a.ThresholdPackage) {
		cfg.Threshold.Package = a.ThresholdPackage
	}

	if !isCIDefaultnt(a.ThresholdPackage) {
		cfg.Threshold.Total = a.ThresholdTotal
	}

	if !isCIDefaultString(a.BadgeFileName) {
		cfg.Badge.FileName = a.BadgeFileName
	}

	if !isCIDefaultString(a.CDNSecret) {
		cfg.Badge.CDN.Secret = a.CDNSecret
		cfg.Badge.CDN.Key = escapeCiDefaultString(a.CDNKey)
		cfg.Badge.CDN.Region = escapeCiDefaultString(a.CDNRegion)
		cfg.Badge.CDN.FileName = escapeCiDefaultString(a.CDNFileName)
		cfg.Badge.CDN.BucketName = escapeCiDefaultString(a.CDNBucketName)
		cfg.Badge.CDN.ForcePathStyle = a.CDNForcePathStyle

		if !isCIDefaultString(a.CDNEndpoint) {
			cfg.Badge.CDN.Endpoint = a.CDNEndpoint
		}
	}

	if !isCIDefaultString(a.GitToken) {
		cfg.Badge.Git.Token = a.GitToken
		cfg.Badge.Git.Repository = escapeCiDefaultString(a.GitRepository)
		cfg.Badge.Git.Branch = escapeCiDefaultString(a.GitBranch)
		cfg.Badge.Git.FileName = escapeCiDefaultString(a.GitFileName)
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
	if !isCIDefaultString(cmdArgs.ConfigPath) {
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

func isCIDefaultString(v string) bool { return v == ciDefaultString }

func isCIDefaultnt(v int) bool { return v == ciDefaultnt }

func escapeCiDefaultString(v string) string {
	if v == ciDefaultString {
		return ""
	}

	return v
}
