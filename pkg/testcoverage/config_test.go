package testcoverage_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	. "github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage"
)

const nonEmptyStr = "any"

func Test_Config_Validate(t *testing.T) {
	t.Parallel()

	cfg := newValidCfg()
	assert.NoError(t, cfg.Validate())

	cfg = newValidCfg()
	cfg.Profile = ""
	assert.ErrorIs(t, cfg.Validate(), ErrCoverageProfileNotSpecified)

	cfg = newValidCfg()
	cfg.Threshold.File = 101
	assert.ErrorIs(t, cfg.Validate(), ErrThresholdNotInRange)

	cfg = newValidCfg()
	cfg.Threshold.File = -1
	assert.ErrorIs(t, cfg.Validate(), ErrThresholdNotInRange)

	cfg = newValidCfg()
	cfg.Threshold.Package = 101
	assert.ErrorIs(t, cfg.Validate(), ErrThresholdNotInRange)

	cfg = newValidCfg()
	cfg.Threshold.Package = -1
	assert.ErrorIs(t, cfg.Validate(), ErrThresholdNotInRange)

	cfg = newValidCfg()
	cfg.Threshold.Total = 101
	assert.ErrorIs(t, cfg.Validate(), ErrThresholdNotInRange)

	cfg = newValidCfg()
	cfg.Threshold.Total = -1
	assert.ErrorIs(t, cfg.Validate(), ErrThresholdNotInRange)

	cfg = newValidCfg()
	cfg.Override = []Override{{Threshold: 101}}
	assert.ErrorIs(t, cfg.Validate(), ErrThresholdNotInRange)

	cfg = newValidCfg()
	cfg.Override = []Override{{Threshold: 100, Path: "("}}
	assert.ErrorIs(t, cfg.Validate(), ErrRegExpNotValid)

	cfg = newValidCfg()
	cfg.Exclude.Paths = []string{"("}
	assert.ErrorIs(t, cfg.Validate(), ErrRegExpNotValid)
}

func Test_Config_ValidateCDN(t *testing.T) {
	t.Parallel()

	cfg := newValidCfg()
	cfg.Badge.CDN.Secret = nonEmptyStr
	assert.ErrorIs(t, cfg.Validate(), ErrCDNOptionNotSet)

	cfg = newValidCfg()
	cfg.Badge.CDN.Secret = nonEmptyStr
	cfg.Badge.CDN.Key = nonEmptyStr
	assert.ErrorIs(t, cfg.Validate(), ErrCDNOptionNotSet)

	cfg = newValidCfg()
	cfg.Badge.CDN.Secret = nonEmptyStr
	cfg.Badge.CDN.Key = nonEmptyStr
	cfg.Badge.CDN.Region = nonEmptyStr
	assert.ErrorIs(t, cfg.Validate(), ErrCDNOptionNotSet)

	cfg = newValidCfg()
	cfg.Badge.CDN.Secret = nonEmptyStr
	cfg.Badge.CDN.Key = nonEmptyStr
	cfg.Badge.CDN.Region = nonEmptyStr
	cfg.Badge.CDN.BucketName = nonEmptyStr
	assert.ErrorIs(t, cfg.Validate(), ErrCDNOptionNotSet)

	cfg = newValidCfg()
	cfg.Badge.CDN.Secret = nonEmptyStr
	cfg.Badge.CDN.Key = nonEmptyStr
	cfg.Badge.CDN.Region = nonEmptyStr
	cfg.Badge.CDN.BucketName = nonEmptyStr
	cfg.Badge.CDN.FileName = nonEmptyStr
	cfg.Badge.CDN.Endpoint = nonEmptyStr
	assert.NoError(t, cfg.Validate())
}

func Test_Config_ValidateGit(t *testing.T) {
	t.Parallel()

	cfg := newValidCfg()
	cfg.Badge.Git.Token = nonEmptyStr
	assert.ErrorIs(t, cfg.Validate(), ErrGitOptionNotSet)

	cfg = newValidCfg()
	cfg.Badge.Git.Token = nonEmptyStr
	cfg.Badge.Git.Owner = nonEmptyStr
	assert.ErrorIs(t, cfg.Validate(), ErrGitOptionNotSet)

	cfg = newValidCfg()
	cfg.Badge.Git.Token = nonEmptyStr
	cfg.Badge.Git.Owner = nonEmptyStr
	cfg.Badge.Git.Repository = nonEmptyStr
	assert.ErrorIs(t, cfg.Validate(), ErrGitOptionNotSet)

	cfg = newValidCfg()
	cfg.Badge.Git.Token = nonEmptyStr
	cfg.Badge.Git.Owner = nonEmptyStr
	cfg.Badge.Git.Repository = nonEmptyStr
	cfg.Badge.Git.Branch = nonEmptyStr
	assert.ErrorIs(t, cfg.Validate(), ErrGitOptionNotSet)

	cfg = newValidCfg()
	cfg.Badge.Git.Token = nonEmptyStr
	cfg.Badge.Git.Owner = nonEmptyStr
	cfg.Badge.Git.Repository = nonEmptyStr
	cfg.Badge.Git.Branch = nonEmptyStr
	cfg.Badge.Git.FileName = nonEmptyStr
	assert.NoError(t, cfg.Validate())
}

func Test_ConfigFromFile(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	setFileWithContent := func(name string, content []byte) {
		f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			t.Errorf("could not open file: %v", err)
		}

		_, err = f.Write(content)
		assert.NoError(t, err)

		assert.NoError(t, f.Close())
	}

	t.Run("no file", func(t *testing.T) {
		t.Parallel()

		cfg := Config{}
		err := ConfigFromFile(&cfg, t.TempDir())
		assert.Error(t, err)
		assert.Equal(t, Config{}, cfg)
	})

	t.Run("invalid file", func(t *testing.T) {
		t.Parallel()

		fileName := t.TempDir() + "file.yml"
		setFileWithContent(fileName, []byte("-----"))

		cfg := Config{}
		err := ConfigFromFile(&cfg, fileName)
		assert.Error(t, err)
		assert.Equal(t, Config{}, cfg)
	})

	t.Run("ok file", func(t *testing.T) {
		t.Parallel()

		savedCfg := nonZeroConfig()
		data, err := yaml.Marshal(savedCfg)
		assert.NoError(t, err)

		fileName := t.TempDir() + "file.yml"
		setFileWithContent(fileName, data)

		cfg := Config{}
		err = ConfigFromFile(&cfg, fileName)
		assert.NoError(t, err)
		assert.Equal(t, savedCfg, cfg)
	})
}

func TestConfigYamlParse(t *testing.T) {
	t.Parallel()

	zeroCfg := nonZeroConfig()
	data, err := yaml.Marshal(zeroCfg)
	assert.NoError(t, err)
	assert.YAMLEq(t, string(data), nonZeroYaml())

	cfg := Config{}
	err = yaml.Unmarshal([]byte(nonZeroYaml()), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, nonZeroConfig(), cfg)
}

func nonZeroConfig() Config {
	return Config{
		Profile:     "cover.out",
		LocalPrefix: "prefix",
		Threshold:   Threshold{100, 100, 100},
		Override:    []Override{{Path: "pathToFile", Threshold: 99}},
		Exclude: Exclude{
			Paths: []string{"path1", "path2"},
		},
		BreakdownFileName: "breakdown.testcoverage",
		Diff: Diff{
			BaseBreakdownFileName: "breakdown.testcoverage",
		},
		GithubActionOutput: true,
	}
}

func nonZeroYaml() string {
	return `
profile: cover.out
local-prefix: prefix
threshold:
    file: 100
    package: 100
    total: 100
override:
    - threshold: 99
      path: pathToFile
exclude:
  paths:
    - path1
    - path2
breakdown-file-name: 'breakdown.testcoverage'
diff:
  base-breakdown-file-name: 'breakdown.testcoverage'	
github-action-output: true`
}

func newValidCfg() Config {
	return Config{Profile: "cover.out"}
}
