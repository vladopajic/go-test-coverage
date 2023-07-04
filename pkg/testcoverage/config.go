package testcoverage

import (
	"fmt"
	"os"
	"regexp"

	yaml "gopkg.in/yaml.v3"
)

var (
	ErrThresholdNotInRange         = fmt.Errorf("threshold must be in range [0 - 100]")
	ErrCoverageProfileNotSpecified = fmt.Errorf("coverage profile file not specified")
	ErrRegExpNotValid              = fmt.Errorf("regular expression is not valid")
)

type Config struct {
	Profile            string     `yaml:"profile"`
	LocalPrefix        string     `yaml:"local-prefix"`
	Threshold          Threshold  `yaml:"threshold"`
	Override           []Override `yaml:"override,omitempty"`
	Exclude            Exclude    `yaml:"exclude"`
	GithubActionOutput bool       `yaml:"github-action-output"`
}

type Threshold struct {
	File    int `yaml:"file"`
	Package int `yaml:"package"`
	Total   int `yaml:"total"`
}

type Override struct {
	Threshold int    `yaml:"threshold"`
	Path      string `yaml:"path"`
}

type Exclude struct {
	Paths []string `yaml:"paths,omitempty"`
}

//nolint:cyclop // relax
func (c Config) Validate() error {
	inRange := func(t int) bool { return t >= 0 && t <= 100 }
	validateRegexp := func(s string) error {
		_, err := regexp.Compile("(?i)" + s)
		return err //nolint:wrapcheck // relax
	}

	if c.Profile == "" {
		return ErrCoverageProfileNotSpecified
	}

	if !inRange(c.Threshold.File) {
		return fmt.Errorf("file %w", ErrThresholdNotInRange)
	}

	if !inRange(c.Threshold.Package) {
		return fmt.Errorf("package %w", ErrThresholdNotInRange)
	}

	if !inRange(c.Threshold.Total) {
		return fmt.Errorf("total %w", ErrThresholdNotInRange)
	}

	for i, pattern := range c.Exclude.Paths {
		if err := validateRegexp(pattern); err != nil {
			return fmt.Errorf("%w for excluded paths element[%d]: %w", ErrRegExpNotValid, i, err)
		}
	}

	for i, o := range c.Override {
		if !inRange(o.Threshold) {
			return fmt.Errorf("override element[%d] %w", i, ErrThresholdNotInRange)
		}

		if err := validateRegexp(o.Path); err != nil {
			return fmt.Errorf("%w for override element[%d]: %w", ErrRegExpNotValid, i, err)
		}
	}

	return nil
}

func ConfigFromFile(cfg *Config, filename string) error {
	source, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed reading file: %w", err)
	}

	err = yaml.Unmarshal(source, cfg)
	if err != nil {
		return fmt.Errorf("failed parsing config file: %w", err)
	}

	return nil
}
