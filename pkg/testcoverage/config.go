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
	Profile            string    `yaml:"profile"`
	LocalPrefix        string    `yaml:"local-prefix"`
	Threshold          Threshold `yaml:"threshold"`
	Exclude            Exclude   `yaml:"exclude"`
	GithubActionOutput bool      `yaml:"github-action-output"`
}

type Threshold struct {
	File    int `yaml:"file"`
	Package int `yaml:"package"`
	Total   int `yaml:"total"`
}

type Exclude struct {
	Paths []string `yaml:"paths,omitempty"`
}

func (c Config) Validate() error {
	inRange := func(t int) bool { return t >= 0 && t <= 100 }

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
		_, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			return fmt.Errorf("%w for paths at position %d: %w", ErrRegExpNotValid, i, err)
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
