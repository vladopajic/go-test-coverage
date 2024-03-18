package testcoverage

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v3"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badgestorer"
)

var (
	ErrThresholdNotInRange         = errors.New("threshold must be in range [0 - 100]")
	ErrCoverageProfileNotSpecified = errors.New("coverage profile file not specified")
	ErrRegExpNotValid              = errors.New("regular expression is not valid")
	ErrCDNOptionNotSet             = errors.New("CDN options are not valid")
	ErrGitOptionNotSet             = errors.New("git options are not valid")
)

type Config struct {
	Profile            string     `yaml:"profile"`
	LocalPrefix        string     `yaml:"local-prefix"`
	SourceDir          string     `yaml:"-"`
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
	CDN      badgestorer.CDN
	Git      badgestorer.Git
}

func (c Config) Validate() error {
	validateRegexp := func(s string) error {
		_, err := regexp.Compile("(?i)" + s)
		return err //nolint:wrapcheck // error is wrapped at level above
	}

	if c.Profile == "" {
		return ErrCoverageProfileNotSpecified
	}

	if err := c.validateThreshold(); err != nil {
		return err
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
		return fmt.Errorf("%w: %s", ErrCDNOptionNotSet, err.Error())
	}

	if err := c.validateGit(); err != nil {
		return fmt.Errorf("%w: %s", ErrGitOptionNotSet, err.Error())
	}

	return nil
}

func (c Config) validateThreshold() error {
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

func (c Config) validateCDN() error {
	// when cnd config is empty, cnd featue is disabled and it's not need to validate
	if reflect.DeepEqual(c.Badge.CDN, badgestorer.CDN{}) {
		return nil
	}

	return hasNonEmptyFields(c.Badge.CDN)
}

func (c Config) validateGit() error {
	// when git config is empty, git featue is disabled and it's not need to validate
	if reflect.DeepEqual(c.Badge.Git, badgestorer.Git{}) {
		return nil
	}

	return hasNonEmptyFields(c.Badge.Git)
}

func hasNonEmptyFields(obj any) error {
	v := reflect.ValueOf(obj)
	for i := range v.NumField() {
		f := v.Field(i)

		if !f.IsZero() { // filed is set
			continue
		}

		if f.Type().Kind() == reflect.Bool { // boolean fields are always set
			continue
		}

		name := strings.ToLower(v.Type().Field(i).Name)

		return fmt.Errorf("property [%v] should be set", name)
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

func inRange(t int) bool { return t >= 0 && t <= 100 }
