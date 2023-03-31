package testcoverage_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/pkg/testcoverage"
)

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
