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
	Debug              *bool   `arg:"-d,--debug"`
	SourceDir          *string `arg:"-s,--source-dir"`
	GithubActionOutput *bool   `arg:"-o,--github-action-output"`
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
	CDNForcePathStyle *bool   `arg:"--cdn-force-path-style"`

	GitToken      *string `arg:"--git-token"`
	GitRepository *string `arg:"--git-repository"`
	GitBranch     *string `arg:"--git-branch"`
	GitFileName   *string `arg:"--git-file-name"`
}

func (*args) Version() string {
	return Name + " " + Version
}

func (a *args) overrideConfig(cfg testcoverage.Config) (testcoverage.Config, error) {
	setValue(&cfg.Profile, a.Profile)
	setValue(&cfg.Debug, a.Debug)
	setValue(&cfg.SourceDir, a.SourceDir)
	setValue(&cfg.GithubActionOutput, a.GithubActionOutput)
	setValue(&cfg.Threshold.File, a.ThresholdFile)
	setValue(&cfg.Threshold.Package, a.ThresholdPackage)
	setValue(&cfg.Threshold.Total, a.ThresholdTotal)

	setValue(&cfg.BreakdownFileName, a.BreakdownFileName)
	setValue(&cfg.Diff.BaseBreakdownFileName, a.DiffBaseBreakdownFileName)

	setValue(&cfg.Badge.FileName, a.BadgeFileName)

	if a.CDNSecret != nil {
		setValue(&cfg.Badge.CDN.Secret, a.CDNSecret)
		setValue(&cfg.Badge.CDN.Key, a.CDNKey)
		setValue(&cfg.Badge.CDN.Region, a.CDNRegion)
		setValue(&cfg.Badge.CDN.FileName, a.CDNFileName)
		setValue(&cfg.Badge.CDN.BucketName, a.CDNBucketName)
		setValue(&cfg.Badge.CDN.ForcePathStyle, a.CDNForcePathStyle)
		setValue(&cfg.Badge.CDN.Endpoint, a.CDNEndpoint)
	}

	if a.GitToken != nil {
		setValue(&cfg.Badge.Git.Token, a.GitToken)
		setValue(&cfg.Badge.Git.Branch, a.GitBranch)
		setValue(&cfg.Badge.Git.FileName, a.GitFileName)

		if a.GitRepository != nil {
			parts := strings.Split(*a.GitRepository, "/")
			if len(parts) != 2 { //nolint:mnd // relax
				return cfg, errors.New("--git-repository flag should have format {owner}/{repository}")
			}

			cfg.Badge.Git.Owner = parts[0]
			cfg.Badge.Git.Repository = parts[1]
		}
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

func setValue[T any](dest *T, source *T) {
	if source != nil {
		*dest = *source
	}
}
