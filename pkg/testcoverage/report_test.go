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

	t.Run("all - pass", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		ReportForHuman(buf, AnalyzeResult{MeetsTotalCoverage: true})
		assertHumanReport(t, buf.String(), 3, 0)
	})

	t.Run("total coverage - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		ReportForHuman(buf, AnalyzeResult{MeetsTotalCoverage: false})
		assertHumanReport(t, buf.String(), 2, 1)
	})

	t.Run("file coverage - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Threshold: Threshold{File: 10}}
		statsWithError := randStats(prefix, 0, 9)
		statsNoError := randStats(prefix, 10, 100)
		result := Analyze(cfg, mergeStats(statsWithError, statsNoError))
		ReportForHuman(buf, result)
		assertHumanReport(t, buf.String(), 2, 1)
		assertContainStats(t, buf.String(), statsWithError)
		assertNotContainStats(t, buf.String(), statsNoError)
	})

	t.Run("package coverage - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Threshold: Threshold{Package: 10}}
		statsWithError := randStats(prefix, 0, 9)
		statsNoError := randStats(prefix, 10, 100)
		result := Analyze(cfg, mergeStats(statsWithError, statsNoError))
		ReportForHuman(buf, result)
		assertHumanReport(t, buf.String(), 2, 1)
		assertContainStats(t, buf.String(), MakePackageStats(statsWithError))
		assertNotContainStats(t, buf.String(), MakePackageStats(statsNoError))
		assertNotContainStats(t, buf.String(), statsWithError)
		assertNotContainStats(t, buf.String(), statsNoError)
	})
}

func Test_ReportForGithubAction(t *testing.T) {
	t.Parallel()

	prefix := "organization.org/pkg/"

	t.Run("total coverage - pass", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Threshold: Threshold{Total: 100}}
		statsNoError := randStats(prefix, 100, 100)
		result := Analyze(cfg, statsNoError)
		ReportForGithubAction(buf, result)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertNotContainStats(t, buf.String(), statsNoError)
	})

	t.Run("total coverage - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		statsWithError := randStats(prefix, 0, 9)
		statsNoError := randStats(prefix, 10, 100)
		cfg := Config{Threshold: Threshold{Total: 10}}
		result := Analyze(cfg, mergeStats(statsWithError, statsNoError))
		ReportForGithubAction(buf, result)
		assertGithubActionErrorsCount(t, buf.String(), 1)
		assertNotContainStats(t, buf.String(), statsWithError)
		assertNotContainStats(t, buf.String(), statsNoError)
	})

	t.Run("file coverage - pass", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Threshold: Threshold{File: 10}}
		statsNoError := randStats(prefix, 10, 100)
		result := Analyze(cfg, statsNoError)
		ReportForGithubAction(buf, result)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertNotContainStats(t, buf.String(), statsNoError)
	})

	t.Run("file coverage - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Threshold: Threshold{File: 10}}
		statsWithError := randStats(prefix, 0, 9)
		statsNoError := randStats(prefix, 10, 100)
		result := Analyze(cfg, mergeStats(statsWithError, statsNoError))
		ReportForGithubAction(buf, result)
		assertGithubActionErrorsCount(t, buf.String(), len(statsWithError))
		assertContainStats(t, buf.String(), statsWithError)
		assertNotContainStats(t, buf.String(), statsNoError)
	})

	t.Run("package coverage - pass", func(t *testing.T) {
		t.Parallel()
		buf := &bytes.Buffer{}
		cfg := Config{Threshold: Threshold{Package: 10}}
		statsNoError := randStats(prefix, 10, 100)
		result := Analyze(cfg, statsNoError)
		ReportForGithubAction(buf, result)
		assertGithubActionErrorsCount(t, buf.String(), 0)
		assertNotContainStats(t, buf.String(), MakePackageStats(statsNoError))
		assertNotContainStats(t, buf.String(), statsNoError)
	})

	t.Run("package coverage - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Threshold: Threshold{Package: 10}}
		statsWithError := randStats(prefix, 0, 9)
		statsNoError := randStats(prefix, 10, 100)
		result := Analyze(cfg, mergeStats(statsWithError, statsNoError))
		ReportForGithubAction(buf, result)
		assertGithubActionErrorsCount(t, buf.String(), len(MakePackageStats(statsWithError)))
		assertContainStats(t, buf.String(), MakePackageStats(statsWithError))
		assertNotContainStats(t, buf.String(), MakePackageStats(statsNoError))
		assertNotContainStats(t, buf.String(), statsWithError)
		assertNotContainStats(t, buf.String(), statsNoError)
	})

	t.Run("file, package and total - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Threshold: Threshold{File: 10, Package: 10, Total: 100}}
		statsWithError := randStats(prefix, 0, 9)
		statsNoError := randStats(prefix, 10, 100)
		totalErrorsCount := len(MakePackageStats(statsWithError)) + len(statsWithError) + 1
		result := Analyze(cfg, mergeStats(statsWithError, statsNoError))
		ReportForGithubAction(buf, result)
		assertGithubActionErrorsCount(t, buf.String(), totalErrorsCount)
		assertContainStats(t, buf.String(), statsWithError)
		assertNotContainStats(t, buf.String(), MakePackageStats(statsNoError))
		assertNotContainStats(t, buf.String(), statsNoError)
	})
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

	colors := make(map[string]struct{})

	{ // Assert that there are 5 colors for coverage [0-101]
		for i := 0; i <= 101; i++ {
			color := CoverageColor(i)
			colors[color] = struct{}{}
		}

		assert.Len(t, colors, 6)
	}

	{ // Assert valid color values
		isHexColor := func(color string) bool {
			return string(color[0]) == "#" && len(color) == 7
		}

		for color := range colors {
			assert.True(t, isHexColor(color))
		}
	}
}
