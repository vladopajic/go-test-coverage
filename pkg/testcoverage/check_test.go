package testcoverage_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
)

const (
	profileOK  = "testdata/ok.profile"
	profileNOK = "testdata/nok.profile"
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
		result, err := Check(buf, Config{})
		assert.Error(t, err)
		assert.Empty(t, result)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 0, 0)
	})

	t.Run("invalid profile", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileNOK, Threshold: Threshold{Total: 65}}
		result, err := Check(buf, cfg)
		assert.Error(t, err)
		assert.False(t, result.Pass())
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 0, 0)
	})

	t.Run("valid profile - pass", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, Threshold: Threshold{Total: 65}}
		result, err := Check(buf, cfg)
		assert.NoError(t, err)
		assert.True(t, result.Pass())
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 3, 0)
	})

	t.Run("valid profile with exclude - pass", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{
			Profile:   profileOK,
			Threshold: Threshold{Total: 100},
			Exclude:   Exclude{Paths: []string{`cover\.go$`}},
		}
		result, err := Check(buf, cfg)
		assert.NoError(t, err)
		assert.True(t, result.Pass())
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 3, 0)
	})

	t.Run("valid profile - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, Threshold: Threshold{Total: 100}}
		result, err := Check(buf, cfg)
		assert.NoError(t, err)
		assert.False(t, result.Pass())
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 2, 1)
		assert.GreaterOrEqual(t, strings.Count(buf.String(), prefix), 0)
	})

	t.Run("valid profile - fail with prefix", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}

		cfg := Config{Profile: profileOK, LocalPrefix: prefix, Threshold: Threshold{Total: 65}}
		result, err := Check(buf, cfg)
		assert.NoError(t, err)
		assert.True(t, result.Pass())
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 3, 0)
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
		result, err := Check(buf, cfg)
		assert.NoError(t, err)
		assert.True(t, result.Pass())
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 3, 0)
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
		result, err := Check(buf, cfg)
		assert.NoError(t, err)
		assert.False(t, result.Pass())
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 1, 2)
		assert.GreaterOrEqual(t, strings.Count(buf.String(), prefix), 0)
	})
}

//nolint:paralleltest // must not be parallel because it uses env
func TestCheckNoParallel(t *testing.T) {
	if testing.Short() {
		return
	}

	t.Run("ok fail; no github output file", func(t *testing.T) {
		t.Setenv(GaOutputFileEnv, "")

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, GithubActionOutput: true, Threshold: Threshold{Total: 100}}
		_, err := Check(buf, cfg)
		assert.Error(t, err)
	})

	t.Run("ok pass; with github output file", func(t *testing.T) {
		testFile := t.TempDir() + "/ga.output"
		t.Setenv(GaOutputFileEnv, testFile)

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, GithubActionOutput: true, Threshold: Threshold{Total: 10}}
		result, err := Check(buf, cfg)
		assert.NoError(t, err)
		assert.True(t, result.Pass())
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertHumanReport(t, buf.String(), 3, 0)
		assertGithubOutputValues(t, testFile)
	})

	t.Run("ok fail; with github output file", func(t *testing.T) {
		testFile := t.TempDir() + "/ga.output"
		t.Setenv(GaOutputFileEnv, testFile)

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, GithubActionOutput: true, Threshold: Threshold{Total: 100}}
		result, err := Check(buf, cfg)
		assert.NoError(t, err)
		assert.False(t, result.Pass())
		assertGithubActionErrorsCount(t, buf.String(), 1)
		assertHumanReport(t, buf.String(), 2, 1)
		assertGithubOutputValues(t, testFile)
	})
}

func Test_Analyze(t *testing.T) {
	t.Parallel()

	prefix := "organization.org/" + randName()

	t.Run("nil coverage stats", func(t *testing.T) {
		t.Parallel()

		result := Analyze(Config{}, nil)
		assert.Empty(t, result.FilesBelowThreshold)
		assert.Empty(t, result.PackagesBelowThreshold)
		assert.Equal(t, 0, result.TotalCoverage)
	})

	t.Run("total coverage above threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{LocalPrefix: prefix, Threshold: Threshold{Total: 10}},
			randStats(prefix, 10, 100),
		)
		assert.True(t, result.Pass())
		assertPrefix(t, result, prefix, false)

		result = Analyze(
			Config{Threshold: Threshold{Total: 10}},
			randStats(prefix, 10, 100),
		)
		assert.True(t, result.Pass())
		assertPrefix(t, result, prefix, true)
	})

	t.Run("total coverage below threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{Threshold: Threshold{Total: 10}},
			randStats(prefix, 0, 9),
		)
		assert.False(t, result.Pass())
	})

	t.Run("files coverage above threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{LocalPrefix: prefix, Threshold: Threshold{File: 10}},
			randStats(prefix, 10, 100),
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
		)
		assert.Empty(t, result.FilesBelowThreshold)
		assert.NotEmpty(t, result.PackagesBelowThreshold)
		assert.False(t, result.Pass())
		assertPrefix(t, result, prefix, true)
	})
}
