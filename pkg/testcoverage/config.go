package testcoverage

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"

	yaml "gopkg.in/yaml.v3"
)

var (
	ErrThresholdNotInRange         = fmt.Errorf("threshold must be in range [0 - 100]")
	ErrCoverageProfileNotSpecified = fmt.Errorf("coverage profile file not specified")
	ErrRegExpNotValid              = fmt.Errorf("regular expression is not valid")
	ErrCDNOptionNotSet             = fmt.Errorf("cdn option not set")
)

type Config struct {
	Profile            string     `yaml:"profile"`
	LocalPrefix        string     `yaml:"local-prefix"`
	Threshold          Threshold  `yaml:"threshold"`
	Override           []Override `yaml:"override,omitempty"`
	Exclude            Exclude    `yaml:"exclude"`
	GithubActionOutput bool       `yaml:"github-action-output"`
	Badge              Badge      `yaml:"-"`
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

type Badge struct {
	FileName string
	CDN      CDN
	Git      Git
}

//nolint:cyclop  // relax
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

	if err := c.validateCDN(); err != nil {
		return fmt.Errorf("%w, %s", ErrCDNOptionNotSet, err.Error())
	}

	return nil
}

//nolint:goerr113,wsl // relax
func (c Config) validateCDN() error {
	// when cnd config is empty, cnd featue is disabled and it's not need to validate
	if reflect.DeepEqual(c.Badge.CDN, CDN{}) {
		return nil
	}

	cdn := c.Badge.CDN

	if cdn.Key == "" {
		return errors.New("CDN key should be set")
	}
	if cdn.Secret == "" {
		return errors.New("CDN secret should be set")
	}
	if cdn.Region == "" {
		return errors.New("CDN region should be set")
	}
	if cdn.BucketName == "" {
		return errors.New("CDN bucket name should be set")
	}
	if cdn.FileName == "" {
		return errors.New("CDN file name should be set")
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
