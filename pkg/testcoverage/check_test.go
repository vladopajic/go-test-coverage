package testcoverage_test

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/path"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/testdata"
)

const (
	testdataDir  = "testdata/"
	profileOK    = testdataDir + testdata.ProfileOK
	profileNOK   = testdataDir + testdata.ProfileNOK
	breakdownOK  = testdataDir + testdata.BreakdownOK
	breakdownNOK = testdataDir + testdata.BreakdownNOK
)

func TestCheck(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	prefix := "github.com/vladopajic/go-test-coverage/v2"

	t.Run("no profile", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		pass := Check(buf, Config{})
		assert.False(t, pass)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 0, 0)
	})

	t.Run("invalid profile", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileNOK, Threshold: Threshold{Total: 65}}
		pass := Check(buf, cfg)
		assert.False(t, pass)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 0, 0)
	})

	t.Run("valid profile - pass", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, Threshold: Threshold{Total: 65}}
		pass := Check(buf, cfg)
		assert.True(t, pass)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 1, 0)
	})

	t.Run("valid profile with exclude - pass", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{
			Profile:   profileOK,
			Threshold: Threshold{Total: 100},
			Exclude: Exclude{
				Paths: []string{`cdn\.go$`, `github\.go$`, `cover\.go$`, `check\.go$`, `path\.go$`},
			},
		}
		pass := Check(buf, cfg)
		assert.True(t, pass)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 1, 0)
	})

	t.Run("valid profile - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, Threshold: Threshold{Total: 100}}
		pass := Check(buf, cfg)
		assert.False(t, pass)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 0, 1)
		assert.GreaterOrEqual(t, strings.Count(buf.String(), prefix), 0)
	})

	t.Run("valid profile - fail with prefix", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}

		cfg := Config{Profile: profileOK, LocalPrefix: prefix, Threshold: Threshold{Total: 65}}
		pass := Check(buf, cfg)
		assert.True(t, pass)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 1, 0)
		assert.Equal(t, 0, strings.Count(buf.String(), prefix))
	})

	t.Run("valid profile - pass after override", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{
			Profile:   profileOK,
			Threshold: Threshold{File: 100},
			Override:  []Override{{Threshold: 10, Path: "^pkg"}},
		}
		pass := Check(buf, cfg)
		assert.True(t, pass)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 2, 0)
		assert.GreaterOrEqual(t, strings.Count(buf.String(), prefix), 0)
	})

	t.Run("valid profile - fail after override", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{
			Profile:   profileOK,
			Threshold: Threshold{File: 10},
			Override:  []Override{{Threshold: 100, Path: "^pkg"}},
		}
		pass := Check(buf, cfg)
		assert.False(t, pass)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 0, 2)
		assert.GreaterOrEqual(t, strings.Count(buf.String(), prefix), 0)
	})

	t.Run("valid profile - fail couldn't save badge", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{
			Profile: profileOK,
			Badge: Badge{
				FileName: t.TempDir(), // should failed because this is dir
			},
		}
		pass := Check(buf, cfg)
		assert.False(t, pass)
		assertFailedToSaveBadge(t, buf.String())
	})

	t.Run("valid profile - fail invalid breakdown file", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{
			Profile:           profileOK,
			BreakdownFileName: t.TempDir(), // should failed because this is dir
		}
		pass := Check(buf, cfg)
		assert.False(t, pass)
		assert.Contains(t, buf.String(), "failed to save coverage breakdown")
	})

	t.Run("valid profile - valid breakdown file", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{
			Profile:           profileOK,
			BreakdownFileName: t.TempDir() + "/breakdown.testcoverage",
		}
		pass := Check(buf, cfg)
		assert.True(t, pass)

		contentBytes, err := os.ReadFile(cfg.BreakdownFileName)
		assert.NoError(t, err)
		assert.NotEmpty(t, contentBytes)

		stats, err := GenerateCoverageStats(cfg)
		assert.NoError(t, err)
		assert.Equal(t, coverage.SerializeStats(stats), contentBytes)
	})

	t.Run("valid profile - invalid base breakdown file", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{
			Profile: profileOK,
			Diff: Diff{
				BaseBreakdownFileName: t.TempDir(), // should failed because this is dir
			},
		}
		pass := Check(buf, cfg)
		assert.False(t, pass)
		assert.Contains(t, buf.String(), "failed to load base coverage breakdown")
	})
}

// must not be parallel because it uses env
func TestCheckNoParallel(t *testing.T) {
	if testing.Short() {
		return
	}

	t.Run("ok fail; no github output file", func(t *testing.T) {
		t.Setenv(GaOutputFileEnv, "")

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, GithubActionOutput: true, Threshold: Threshold{Total: 100}}
		pass := Check(buf, cfg)
		assert.False(t, pass)
	})

	t.Run("ok pass; with github output file", func(t *testing.T) {
		testFile := t.TempDir() + "/ga.output"
		t.Setenv(GaOutputFileEnv, testFile)

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, GithubActionOutput: true, Threshold: Threshold{Total: 10}}
		pass := Check(buf, cfg)
		assert.True(t, pass)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 1, 0)
		assertGithubOutputValues(t, testFile)
	})

	t.Run("ok fail; with github output file", func(t *testing.T) {
		testFile := t.TempDir() + "/ga.output"
		t.Setenv(GaOutputFileEnv, testFile)

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, GithubActionOutput: true, Threshold: Threshold{Total: 100}}
		pass := Check(buf, cfg)
		assert.False(t, pass)
		assertGithubActionErrorsCount(t, buf.String(), 1)
		assertHumanReport(t, buf.String(), 0, 1)
		assertGithubOutputValues(t, testFile)
	})
}

func Test_Analyze(t *testing.T) {
	t.Parallel()

	prefix := "organization.org/" + randName()

	t.Run("nil coverage stats", func(t *testing.T) {
		t.Parallel()

		result := Analyze(Config{}, nil, nil)
		assert.Empty(t, result.FilesBelowThreshold)
		assert.Empty(t, result.PackagesBelowThreshold)
		assert.Equal(t, 0, result.TotalStats.CoveredPercentage())
	})

	t.Run("total coverage above threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{LocalPrefix: prefix, Threshold: Threshold{Total: 10}},
			randStats(prefix, 10, 100),
			nil,
		)
		assert.True(t, result.Pass())
		assertPrefix(t, result, prefix, false)

		result = Analyze(
			Config{Threshold: Threshold{Total: 10}},
			randStats(prefix, 10, 100),
			nil,
		)
		assert.True(t, result.Pass())
		assertPrefix(t, result, prefix, true)
	})

	t.Run("total coverage below threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{Threshold: Threshold{Total: 10}},
			randStats(prefix, 0, 9),
			nil,
		)
		assert.False(t, result.Pass())
	})

	t.Run("files coverage above threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{LocalPrefix: prefix, Threshold: Threshold{File: 10}},
			randStats(prefix, 10, 100),
			nil,
		)
		assert.True(t, result.Pass())
		assertPrefix(t, result, prefix, false)
	})

	t.Run("files coverage below threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{Threshold: Threshold{File: 10}},
			mergeStats(
				randStats(prefix, 0, 9),
				randStats(prefix, 10, 100),
			),
			nil,
		)
		assert.NotEmpty(t, result.FilesBelowThreshold)
		assert.Empty(t, result.PackagesBelowThreshold)
		assert.False(t, result.Pass())
		assertPrefix(t, result, prefix, true)
	})

	t.Run("package coverage above threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{LocalPrefix: prefix, Threshold: Threshold{Package: 10}},
			randStats(prefix, 10, 100),
			nil,
		)
		assert.True(t, result.Pass())
		assertPrefix(t, result, prefix, false)
	})

	t.Run("package coverage below threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{Threshold: Threshold{Package: 10}},
			mergeStats(
				randStats(prefix, 0, 9),
				randStats(prefix, 10, 100),
			),
			nil,
		)
		assert.Empty(t, result.FilesBelowThreshold)
		assert.NotEmpty(t, result.PackagesBelowThreshold)
		assert.False(t, result.Pass())
		assertPrefix(t, result, prefix, true)
	})
}

func TestLoadBaseCoverageBreakdown(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	if runtime.GOOS == "windows" {
		t.Skip("tests fails windows in ci, but works locally")
	}

	stats, err := LoadBaseCoverageBreakdown(Config{Diff: Diff{}})
	assert.NoError(t, err)
	assert.Empty(t, stats)

	stats, err = LoadBaseCoverageBreakdown(Config{
		Diff: Diff{BaseBreakdownFileName: path.NormalizeForOS(breakdownOK)},
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 14)

	stats, err = LoadBaseCoverageBreakdown(Config{
		Diff: Diff{BaseBreakdownFileName: t.TempDir()},
	})
	assert.Error(t, err)
	assert.Empty(t, stats)

	stats, err = LoadBaseCoverageBreakdown(Config{
		Diff: Diff{BaseBreakdownFileName: path.NormalizeForOS(breakdownNOK)},
	})
	assert.Error(t, err)
	assert.Empty(t, stats)
}
