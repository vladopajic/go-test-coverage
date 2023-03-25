package testcoverage

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"
)

var (
	ErrThresholdNotInRange         = fmt.Errorf("threshold must be in range [0 - 100]")
	ErrCoverageProfileNotSpecified = fmt.Errorf("coverage profile file not specified")
)

const (
	defaultFileThreshold    = 0
	defaultPackageThreshold = 0
	defaultTotalThreshold   = 50
)

type Config struct {
	Profile     string    `yaml:"profile"`
	LocalPrefix string    `yaml:"localPrefix"`
	Threshold   Threshold `yaml:"threshold"`
}

type Threshold struct {
	File    int `yaml:"file"`
	Package int `yaml:"package"`
	Total   int `yaml:"total"`
}

func NewConfig() Config {
	return Config{
		Threshold: Threshold{
			File:    defaultFileThreshold,
			Package: defaultPackageThreshold,
			Total:   defaultTotalThreshold,
		},
	}
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

	return nil
}

func ConfigFromFile(filename string) (*Config, error) {
	cfg := NewConfig()

	source, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed reading file: %w", err)
	}

	err = yaml.Unmarshal(source, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed parsing config file: %w", err)
	}

	return &cfg, nil
}
