package testcoverage_test

import (
	"bytes"
	"io/ioutil"
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
	assertHumanReport(t, buf.Bytes(), 3, 0)

	// Total coverage error
	buf = &bytes.Buffer{}
	ReportForHuman(buf, AnalyzeResult{MeetsTotalCoverage: false}, Config{})
	assertHumanReport(t, buf.Bytes(), 2, 1)

	// File coverage error
	buf = &bytes.Buffer{}
	result := Analyze(
		Config{LocalPrefix: localPrefix, Threshold: Threshold{File: 10}},
		mergeCoverageStats(
			makeCoverageStats(localPrefix, 9),
			makeCoverageStats(localPrefix, 10),
		),
	)
	ReportForHuman(buf, result, Config{})
	assertHumanReport(t, buf.Bytes(), 2, 1)

	// Package coverage error
	buf = &bytes.Buffer{}
	result = Analyze(
		Config{LocalPrefix: localPrefix, Threshold: Threshold{Package: 10}},
		mergeCoverageStats(
			makeCoverageStats(localPrefix, 9),
			makeCoverageStats(localPrefix, 10),
		),
	)
	ReportForHuman(buf, result, Config{})
	assertHumanReport(t, buf.Bytes(), 2, 1)
}

func assertHumanReport(t *testing.T, output []byte, passCount, failCount int) {
	t.Helper()

	outputStr := string(output)

	assert.Equal(t, passCount, strings.Count(outputStr, "PASS"))
	assert.Equal(t, failCount, strings.Count(outputStr, "FAIL"))
}

func Test_ReportForGithubAction(t *testing.T) {
	t.Parallel()

	localPrefix := "organization.org/" + randName()

	// No errors
	buf := &bytes.Buffer{}
	ReportForGithubAction(buf, AnalyzeResult{MeetsTotalCoverage: true}, Config{})
	assert.Empty(t, buf.Bytes())

	// Total coverage error
	buf = &bytes.Buffer{}
	ReportForGithubAction(buf, AnalyzeResult{MeetsTotalCoverage: false}, Config{})
	assert.NotEmpty(t, buf.Bytes())

	// File coverage error
	buf = &bytes.Buffer{}
	result := Analyze(
		Config{LocalPrefix: localPrefix, Threshold: Threshold{File: 10}},
		mergeCoverageStats(
			makeCoverageStats(localPrefix, 9),
			makeCoverageStats(localPrefix, 10),
		),
	)
	ReportForGithubAction(buf, result, Config{})
	assert.NotEmpty(t, buf.Bytes())

	// Package coverage error
	buf = &bytes.Buffer{}
	result = Analyze(
		Config{LocalPrefix: localPrefix, Threshold: Threshold{Package: 10}},
		mergeCoverageStats(
			makeCoverageStats(localPrefix, 9),
			makeCoverageStats(localPrefix, 10),
		),
	)
	ReportForGithubAction(buf, result, Config{})
	assert.NotEmpty(t, buf.Bytes())
}

func Test_SetGithubActionOutput(t *testing.T) {
	t.Parallel()

	// When test is execute in Github workflow GITHUB_OUTPUT env value will be set.
	// It necessary to preserve this value after test has ended.
	defaultFileVal := os.Getenv(GaOutputFileEnv)
	defer func() {
		err := os.Setenv(GaOutputFileEnv, defaultFileVal)
		assert.NoError(t, err)
	}()

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

		err = SetGithubActionOutput(AnalyzeResult{TotalCoverage: 100})
		assert.NoError(t, err)

		contentBytes, err := ioutil.ReadFile(testFile)
		assert.NoError(t, err)

		content := string(contentBytes)
		assert.Equal(t, 1, strings.Count(content, GaOutputTotalCoverage))
		assert.Equal(t, 1, strings.Count(content, GaOutputBadgeColor))
		assert.Equal(t, 1, strings.Count(content, GaOutputBadgeText))
	}
}
