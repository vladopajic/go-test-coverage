package testcoverage_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/pkg/testcoverage"
)

func Test_ReportForHumann(t *testing.T) {
	t.Parallel()

	localPrefix := "organization.org/" + randName()

	// No errors
	buf := &bytes.Buffer{}
	ReportForHuman(buf, AnalyzeResult{MeetsTotalCoverage: true}, Config{})
	assertHumanReport(t, buf.String(), 3, 0)

	// Total coverage error
	buf = &bytes.Buffer{}
	ReportForHuman(buf, AnalyzeResult{MeetsTotalCoverage: false}, Config{})
	assertHumanReport(t, buf.String(), 2, 1)

	// File coverage error
	buf = &bytes.Buffer{}
	cfg := Config{Threshold: Threshold{File: 10}}
	statsWithError := makeCoverageStats(localPrefix, 0, 9)
	result := Analyze(
		cfg,
		mergeCoverageStats(
			statsWithError,
			makeCoverageStats(localPrefix, 10, 100),
		),
	)
	ReportForHuman(buf, result, cfg)
	assertHumanReport(t, buf.String(), 2, 1)
	assertContainsStatNames(t, buf.String(), statsWithError)

	// Package coverage error
	buf = &bytes.Buffer{}
	cfg = Config{Threshold: Threshold{Package: 10}}
	statsWithError = makeCoverageStats(localPrefix, 0, 9)
	result = Analyze(
		cfg,
		mergeCoverageStats(
			statsWithError,
			makeCoverageStats(localPrefix, 10, 100),
		),
	)
	ReportForHuman(buf, result, cfg)
	assertHumanReport(t, buf.String(), 2, 1)
}

func assertHumanReport(t *testing.T, content string, passCount, failCount int) {
	t.Helper()

	assert.Equal(t, passCount, strings.Count(content, "PASS"))
	assert.Equal(t, failCount, strings.Count(content, "FAIL"))
}

func assertContainsStatNames(t *testing.T, content string, stats []CoverageStats) {
	t.Helper()

	for _, stat := range stats {
		assert.Equal(t, 1, strings.Count(content, stat.Name))
	}
}

func Test_ReportForGithubAction(t *testing.T) {
	t.Parallel()

	localPrefix := "organization.org/" + randName()

	// No errors
	buf := &bytes.Buffer{}
	ReportForGithubAction(buf, AnalyzeResult{MeetsTotalCoverage: true}, Config{})
	assert.Empty(t, buf.Bytes())
	assertGithubActionErrorsCount(t, buf.String(), 0)

	// Total coverage error
	buf = &bytes.Buffer{}
	ReportForGithubAction(buf, AnalyzeResult{MeetsTotalCoverage: false}, Config{})
	assertGithubActionErrorsCount(t, buf.String(), 1)

	// File coverage error
	buf = &bytes.Buffer{}
	cfg := Config{Threshold: Threshold{File: 10}}
	statsWithError := makeCoverageStats(localPrefix, 0, 9)
	result := Analyze(
		cfg,
		mergeCoverageStats(
			statsWithError,
			makeCoverageStats(localPrefix, 10, 100),
		),
	)
	ReportForGithubAction(buf, result, cfg)
	assertGithubActionErrorsCount(t, buf.String(), len(statsWithError))
	assertContainsStatNames(t, buf.String(), statsWithError)

	// Package coverage error
	buf = &bytes.Buffer{}
	cfg = Config{Threshold: Threshold{Package: 10}}
	statsWithError = makeCoverageStats(localPrefix, 0, 9)
	result = Analyze(
		cfg,
		mergeCoverageStats(
			statsWithError,
			makeCoverageStats(localPrefix, 10, 100),
		),
	)
	ReportForGithubAction(buf, result, cfg)
	// assertGithubActionErrorsCount(t, buf.String(), len(MakePackageStats(statsWithError)))
	// assertContainsStatNames(t, buf.String(), MakePackageStats(statsWithError))

	// Total coverage error
	buf = &bytes.Buffer{}
	cfg = Config{Threshold: Threshold{Total: 10}}
	result = Analyze(
		cfg,
		makeCoverageStats(localPrefix, 0, 9),
	)
	ReportForGithubAction(buf, result, Config{})
	assertGithubActionErrorsCount(t, buf.String(), 1)
}

func assertGithubActionErrorsCount(t *testing.T, content string, count int) {
	t.Helper()

	assert.Equal(t, count, strings.Count(content, "::error"))
}

func Test_SetGithubActionOutput(t *testing.T) {
	t.Parallel()

	// When test is execute in Github workflow GITHUB_OUTPUT env value will be set.
	// It necessary to preserve this value after test has ended.
	defaultFileVal := os.Getenv(GaOutputFileEnv)
	defer assert.NoError(t, os.Setenv(GaOutputFileEnv, defaultFileVal))

	{ // Assert case when file is not set in env
		err := os.Setenv(GaOutputFileEnv, "")
		assert.NoError(t, err)

		err = SetGithubActionOutput(AnalyzeResult{})
		assert.Error(t, err)
	}

	{ // Assert case when file is set
		testFile := t.TempDir() + "/ga.output"

		err := os.Setenv(GaOutputFileEnv, testFile)
		assert.NoError(t, err)

		err = SetGithubActionOutput(AnalyzeResult{})
		assert.NoError(t, err)

		contentBytes, err := os.ReadFile(testFile)
		assert.NoError(t, err)

		content := string(contentBytes)
		assert.Equal(t, 1, strings.Count(content, GaOutputTotalCoverage))
		assert.Equal(t, 1, strings.Count(content, GaOutputBadgeColor))
		assert.Equal(t, 1, strings.Count(content, GaOutputBadgeText))
	}
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
