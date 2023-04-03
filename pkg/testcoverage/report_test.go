package testcoverage_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/pkg/testcoverage"
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
	// assertContainStats(t, buf.String(), MakePackageStats(statsWithError))
	assertNotContainStats(t, buf.String(), MakePackageStats(statsNoError))
	assertNotContainStats(t, buf.String(), statsWithError)
	assertNotContainStats(t, buf.String(), statsNoError)
}

func Test_ReportForGithubAction(t *testing.T) {
	t.Parallel()

	prefix := "organization.org"

	// No errors
	buf := &bytes.Buffer{}
	ReportForGithubAction(buf, AnalyzeResult{MeetsTotalCoverage: true}, Threshold{})
	assert.Empty(t, buf.Bytes())
	assertGithubActionErrorsCount(t, buf.String(), 0)

	// Total coverage error
	buf = &bytes.Buffer{}
	ReportForGithubAction(buf, AnalyzeResult{MeetsTotalCoverage: false}, Threshold{})
	assertGithubActionErrorsCount(t, buf.String(), 1)

	// Total coverage error
	buf = &bytes.Buffer{}
	statsWithError := randStats(prefix, 0, 9)
	statsNoError := randStats(prefix, 10, 100)
	cfg := Config{Threshold: Threshold{Total: 10}}
	result := Analyze(cfg, mergeStats(statsWithError, statsNoError))
	ReportForGithubAction(buf, result, cfg.Threshold)
	assertGithubActionErrorsCount(t, buf.String(), 1)
	assertNotContainStats(t, buf.String(), statsWithError)
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

	// Package coverage error
	buf = &bytes.Buffer{}
	cfg = Config{Threshold: Threshold{Package: 10}}
	statsWithError = randStats(prefix, 0, 9)
	statsNoError = randStats(prefix, 10, 100)
	result = Analyze(cfg, mergeStats(statsWithError, statsNoError))
	ReportForGithubAction(buf, result, cfg.Threshold)
	// assertGithubActionErrorsCount(t, buf.String(), len(MakePackageStats(statsWithError)))
	// assertContainStats(t, buf.String(), MakePackageStats(statsWithError))
	assertNotContainStats(t, buf.String(), MakePackageStats(statsNoError))
	assertNotContainStats(t, buf.String(), statsWithError)
	assertNotContainStats(t, buf.String(), statsNoError)
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

func assertHumanReport(t *testing.T, content string, passCount, failCount int) {
	t.Helper()

	assert.Equal(t, passCount, strings.Count(content, "PASS"))
	assert.Equal(t, failCount, strings.Count(content, "FAIL"))
}

func assertContainStats(t *testing.T, content string, stats []CoverageStats) {
	t.Helper()

	contains := 0

	for _, stat := range stats {
		if strings.Count(content, stat.Name) == 1 {
			contains++
		}
	}

	if contains != len(stats) {
		t.Errorf("content doesn't contain exactly one stats: got %d, want %d", contains, len(stats))
	}
}

func assertNotContainStats(t *testing.T, content string, stats []CoverageStats) {
	t.Helper()

	contains := 0

	for _, stat := range stats {
		if strings.Count(content, stat.Name) >= 0 {
			contains++
		}
	}

	if contains != len(stats) {
		t.Errorf("content should not contain stats: got %d", contains)
	}
}

func assertGithubActionErrorsCount(t *testing.T, content string, count int) {
	t.Helper()

	assert.Equal(t, count, strings.Count(content, "::error"))
}
