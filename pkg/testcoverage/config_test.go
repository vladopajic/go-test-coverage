package testcoverage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/pkg/testcoverage"
)

func Test_NewConfig(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()

	assert.Empty(t, cfg.Profile)
	assert.Empty(t, cfg.LocalPrefix)
	assert.False(t, cfg.GithubActionOutput)
	assert.Equal(t, DefaultFileThreshold, cfg.Threshold.File)
	assert.Equal(t, DefaultPackageThreshold, cfg.Threshold.Package)
	assert.Equal(t, DefaultTotalThreshold, cfg.Threshold.Total)
}

func Test_Config_Validate(t *testing.T) {
	t.Parallel()

	newValidCfg := func() Config {
		cfg := NewConfig()
		cfg.Profile = "cover.out"

		return cfg
	}

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
}

func Test_ConfigFromFile(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()
	cfgBefore := NewConfig()
	err := ConfigFromFile(&cfg, t.TempDir())
	assert.Error(t, err)
	assert.Equal(t, cfgBefore, cfg)
}
