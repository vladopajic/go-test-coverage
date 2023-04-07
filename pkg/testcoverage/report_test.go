package testcoverage_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
)

func Test_ReportForHuman(t *testing.T) {
	t.Parallel()

	prefix := "organization.org"

	// No errors
	buf := &bytes.Buffer{}
	ReportForHuman(buf, AnalyzeResult{MeetsTotalCoverage: true}, Threshold{})
	assertHumanReport(t, buf.String(), 3, 0)

	// Total coverage error
	buf = &bytes.Buffer{}
	ReportForHuman(buf, AnalyzeResult{MeetsTotalCoverage: false}, Threshold{})
	assertHumanReport(t, buf.String(), 2, 1)

	// File coverage error
	buf = &bytes.Buffer{}
	cfg := Config{Threshold: Threshold{File: 10}}
	statsWithError := randStats(prefix, 0, 9)
	statsNoError := randStats(prefix, 10, 100)
	result := Analyze(cfg, mergeStats(statsWithError, statsNoError))
	ReportForHuman(buf, result, cfg.Threshold)
	assertHumanReport(t, buf.String(), 2, 1)
	assertContainStats(t, buf.String(), statsWithError)
	assertNotContainStats(t, buf.String(), statsNoError)

	// Package coverage error
	buf = &bytes.Buffer{}
	cfg = Config{Threshold: Threshold{Package: 10}}
	statsWithError = randStats(prefix, 0, 9)
	statsNoError = randStats(prefix, 10, 100)
	result = Analyze(cfg, mergeStats(statsWithError, statsNoError))
	ReportForHuman(buf, result, cfg.Threshold)
	assertHumanReport(t, buf.String(), 2, 1)
	assertContainStats(t, buf.String(), MakePackageStats(statsWithError))
	assertNotContainStats(t, buf.String(), MakePackageStats(statsNoError))
	assertNotContainStats(t, buf.String(), statsWithError)
	assertNotContainStats(t, buf.String(), statsNoError)
}

func Test_ReportForGithubAction(t *testing.T) {
	t.Parallel()

	prefix := "organization.org"

	// Total coverage ok
	buf := &bytes.Buffer{}
	cfg := Config{Threshold: Threshold{Total: 100}}
	statsNoError := randStats(prefix, 100, 100)
	result := Analyze(cfg, statsNoError)
	ReportForGithubAction(buf, result, cfg.Threshold)
	assertGithubActionErrorsCount(t, buf.String(), 0)
	assertNotContainStats(t, buf.String(), statsNoError)

	// Total coverage error
	buf = &bytes.Buffer{}
	statsWithError := randStats(prefix, 0, 9)
	statsNoError = randStats(prefix, 10, 100)
	cfg = Config{Threshold: Threshold{Total: 10}}
	result = Analyze(cfg, mergeStats(statsWithError, statsNoError))
	ReportForGithubAction(buf, result, cfg.Threshold)
	assertGithubActionErrorsCount(t, buf.String(), 1)
	assertNotContainStats(t, buf.String(), statsWithError)
	assertNotContainStats(t, buf.String(), statsNoError)

	// File coverage ok
	buf = &bytes.Buffer{}
	cfg = Config{Threshold: Threshold{File: 10}}
	statsNoError = randStats(prefix, 10, 100)
	result = Analyze(cfg, statsNoError)
	ReportForGithubAction(buf, result, cfg.Threshold)
	assertGithubActionErrorsCount(t, buf.String(), 0)
	assertNotContainStats(t, buf.String(), statsNoError)

	// File coverage error
	buf = &bytes.Buffer{}
	cfg = Config{Threshold: Threshold{File: 10}}
	statsWithError = randStats(prefix, 0, 9)
	statsNoError = randStats(prefix, 10, 100)
	result = Analyze(cfg, mergeStats(statsWithError, statsNoError))
	ReportForGithubAction(buf, result, cfg.Threshold)
	assertGithubActionErrorsCount(t, buf.String(), len(statsWithError))
	assertContainStats(t, buf.String(), statsWithError)
	assertNotContainStats(t, buf.String(), statsNoError)

	// Package coverage ok
	buf = &bytes.Buffer{}
	cfg = Config{Threshold: Threshold{Package: 10}}
	statsNoError = randStats(prefix, 10, 100)
	result = Analyze(cfg, statsNoError)
	ReportForGithubAction(buf, result, cfg.Threshold)
	assertGithubActionErrorsCount(t, buf.String(), 0)
	assertNotContainStats(t, buf.String(), MakePackageStats(statsNoError))
	assertNotContainStats(t, buf.String(), statsNoError)

	// Package coverage error
	buf = &bytes.Buffer{}
	cfg = Config{Threshold: Threshold{Package: 10}}
	statsWithError = randStats(prefix, 0, 9)
	statsNoError = randStats(prefix, 10, 100)
	result = Analyze(cfg, mergeStats(statsWithError, statsNoError))
	ReportForGithubAction(buf, result, cfg.Threshold)
	assertGithubActionErrorsCount(t, buf.String(), len(MakePackageStats(statsWithError)))
	assertContainStats(t, buf.String(), MakePackageStats(statsWithError))
	assertNotContainStats(t, buf.String(), MakePackageStats(statsNoError))
	assertNotContainStats(t, buf.String(), statsWithError)
	assertNotContainStats(t, buf.String(), statsNoError)

	// All below threshold
	buf = &bytes.Buffer{}
	cfg = Config{Threshold: Threshold{File: 10, Package: 10}}
	statsWithError = randStats(prefix, 0, 9)
	statsNoError = randStats(prefix, 10, 100)
	result = Analyze(cfg, mergeStats(statsWithError, statsNoError))
	ReportForGithubAction(buf, result, cfg.Threshold)
	assertGithubActionErrorsCount(t, buf.String(), len(MakePackageStats(statsWithError))+len(statsWithError))
	assertContainStats(t, buf.String(), statsWithError)
	assertNotContainStats(t, buf.String(), MakePackageStats(statsNoError))
	assertNotContainStats(t, buf.String(), statsNoError)
}

//nolint:paralleltest // must not be parallel because it uses env
func Test_SetGithubActionOutput(t *testing.T) {
	if testing.Short() {
		return
	}

	t.Run("no env file", func(t *testing.T) {
		t.Setenv(GaOutputFileEnv, "")

		err := SetGithubActionOutput(AnalyzeResult{})
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		testFile := t.TempDir() + "/ga.output"

		t.Setenv(GaOutputFileEnv, testFile)

		err := SetGithubActionOutput(AnalyzeResult{})
		assert.NoError(t, err)

		contentBytes, err := os.ReadFile(testFile)
		assert.NoError(t, err)

		content := string(contentBytes)
		assert.Equal(t, 1, strings.Count(content, GaOutputTotalCoverage))
		assert.Equal(t, 1, strings.Count(content, GaOutputBadgeColor))
		assert.Equal(t, 1, strings.Count(content, GaOutputBadgeText))
	})
}

func Test_CoverageColor(t *testing.T) {
	t.Parallel()

	{ // Assert that there are 5 colors for coverage [0-101]
		colors := make(map[string]struct{})
		for i := 0; i <= 101; i++ {
			color := CoverageColor(i)
			colors[color] = struct{}{}
		}

		assert.Len(t, colors, 5)
	}
}
