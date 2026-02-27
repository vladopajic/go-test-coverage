package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alexflint/go-arg"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
)

type args struct {
	ConfigPath         *string `arg:"-c,--config"`
	Profile            *string `arg:"-p,--profile"              help:"path to coverage profile"`
	Debug              bool    `arg:"-d,--debug"`
	LocalPrefix        *string `arg:"-l,--local-prefix"` // deprecated
	SourceDir          *string `arg:"-s,--source-dir"`
	GithubActionOutput bool    `arg:"-o,--github-action-output"`
	ThresholdFile      *int    `arg:"-f,--threshold-file"`
	ThresholdPackage   *int    `arg:"-k,--threshold-package"`
	ThresholdTotal     *int    `arg:"-t,--threshold-total"`

	BreakdownFileName         *string `arg:"--breakdown-file-name"`
	DiffBaseBreakdownFileName *string `arg:"--diff-base-breakdown-file-name"`

	BadgeFileName *string `arg:"-b,--badge-file-name"`

	CDNKey            *string `arg:"--cdn-key"`
	CDNSecret         *string `arg:"--cdn-secret"`
	CDNRegion         *string `arg:"--cdn-region"`
	CDNEndpoint       *string `arg:"--cdn-endpoint"`
	CDNFileName       *string `arg:"--cdn-file-name"`
	CDNBucketName     *string `arg:"--cdn-bucket-name"`
	CDNForcePathStyle bool    `arg:"--cdn-force-path-style"`

	GitToken      *string `arg:"--git-token"`
	GitRepository *string `arg:"--git-repository"`
	GitBranch     *string `arg:"--git-branch"`
	GitFileName   *string `arg:"--git-file-name"`
}

func (*args) Version() string {
	return Name + " " + Version
}

//nolint:cyclop,maintidx,mnd,funlen // relax
func (a *args) overrideConfig(cfg testcoverage.Config) (testcoverage.Config, error) {
	if a.Profile != nil {
		cfg.Profile = *a.Profile
	}

	if a.Debug {
		cfg.Debug = true
	}

	if a.GithubActionOutput {
		cfg.GithubActionOutput = true
	}

	if a.LocalPrefix != nil {
		cfg.LocalPrefixDeprecated = *a.LocalPrefix
	}

	if a.SourceDir != nil {
		cfg.SourceDir = *a.SourceDir
	}

	if a.ThresholdFile != nil {
		cfg.Threshold.File = *a.ThresholdFile
	}

	if a.ThresholdPackage != nil {
		cfg.Threshold.Package = *a.ThresholdPackage
	}

	if a.ThresholdTotal != nil {
		cfg.Threshold.Total = *a.ThresholdTotal
	}

	if a.BreakdownFileName != nil {
		cfg.BreakdownFileName = *a.BreakdownFileName
	}

	if a.DiffBaseBreakdownFileName != nil {
		cfg.Diff.BaseBreakdownFileName = *a.DiffBaseBreakdownFileName
	}

	if a.BadgeFileName != nil {
		cfg.Badge.FileName = *a.BadgeFileName
	}

	if a.CDNSecret != nil {
		cfg.Badge.CDN.Secret = *a.CDNSecret
		cfg.Badge.CDN.Key = escapeCiDefaultString(a.CDNKey)
		cfg.Badge.CDN.Region = escapeCiDefaultString(a.CDNRegion)
		cfg.Badge.CDN.FileName = escapeCiDefaultString(a.CDNFileName)
		cfg.Badge.CDN.BucketName = escapeCiDefaultString(a.CDNBucketName)
		cfg.Badge.CDN.ForcePathStyle = a.CDNForcePathStyle

		if a.CDNEndpoint != nil {
			cfg.Badge.CDN.Endpoint = *a.CDNEndpoint
		}
	}

	if a.GitToken != nil {
		cfg.Badge.Git.Token = *a.GitToken
		cfg.Badge.Git.Branch = escapeCiDefaultString(a.GitBranch)
		cfg.Badge.Git.FileName = escapeCiDefaultString(a.GitFileName)

		parts := strings.Split(escapeCiDefaultString(a.GitRepository), "/")
		if len(parts) != 2 {
			return cfg, errors.New("--git-repository flag should have format {owner}/{repository}")
		}

		cfg.Badge.Git.Owner = parts[0]
		cfg.Badge.Git.Repository = parts[1]
	}

	return cfg, nil
}

func readConfig() (testcoverage.Config, error) {
	cmdArgs := &args{}
	arg.MustParse(cmdArgs)

	cfg := testcoverage.Config{}

	// Load config from file
	if cmdArgs.ConfigPath != nil {
		err := testcoverage.ConfigFromFile(&cfg, *cmdArgs.ConfigPath)
		if err != nil {
			return testcoverage.Config{}, fmt.Errorf("failed loading config from file: %w", err)
		}
	}

	// Override config with values from args
	cfg, err := cmdArgs.overrideConfig(cfg)
	if err != nil {
		return testcoverage.Config{}, fmt.Errorf("argument is not valid: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return testcoverage.Config{}, fmt.Errorf("config file is not valid: %w", err)
	}

	return cfg, nil
}

func escapeCiDefaultString(str *string) string {
	if str == nil {
		return ""
	}

	return *str
}
