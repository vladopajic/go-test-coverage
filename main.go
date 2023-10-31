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

const (
	magicString = `''`
	magicInt    = -1
)

func newArgs() args {
	return args{
		ConfigPath:         magicString,
		Profile:            magicString,
		LocalPrefix:        magicString,
		GithubActionOutput: false,
		ThresholdFile:      magicInt,
		ThresholdPackage:   magicInt,
		ThresholdTotal:     magicInt,
		BadgeFileName:      magicString,

		CDNKey:            magicString,
		CDNSecret:         magicString,
		CDNRegion:         magicString,
		CDNEndpoint:       magicString,
		CDNFileName:       magicString,
		CDNBucketName:     magicString,
		CDNForcePathStyle: false,

		GitToken:      magicString,
		GitRepository: magicString,
		GitBranch:     magicString,
		GitFileName:   magicString,
	}
}

func (args) Version() string {
	return "go-test-coverage " + Version
}

//nolint:cyclop // relax
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

	if !isMagicString(a.BadgeFileName) {
		cfg.Badge.FileName = a.BadgeFileName
	}

	if !isMagicString(a.CDNSecret) {
		cfg.Badge.CDN.Secret = a.CDNSecret
		cfg.Badge.CDN.Key = escapeMagicString(a.CDNKey)
		cfg.Badge.CDN.Region = escapeMagicString(a.CDNRegion)
		cfg.Badge.CDN.FileName = escapeMagicString(a.CDNFileName)
		cfg.Badge.CDN.BucketName = escapeMagicString(a.CDNBucketName)
		cfg.Badge.CDN.ForcePathStyle = a.CDNForcePathStyle

		if !isMagicString(a.CDNEndpoint) {
			cfg.Badge.CDN.Endpoint = a.CDNEndpoint
		}
	}

	if !isMagicString(a.GitToken) {
		cfg.Badge.Git.Token = a.GitToken
		cfg.Badge.Git.Repository = escapeMagicString(a.GitRepository)
		cfg.Badge.Git.Branch = escapeMagicString(a.GitBranch)
		cfg.Badge.Git.FileName = escapeMagicString(a.GitFileName)
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
	return v == magicString
}

func isMagicInt(v int) bool {
	return v == magicInt
}

func escapeMagicString(v string) string {
	if v == magicString {
		return ""
	}
	return v
}
