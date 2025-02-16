package testcoverage_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
)

func Test_ReportForHuman(t *testing.T) {
	t.Parallel()

	const prefix = "organization.org"

	thr := Threshold{100, 100, 100}

	t.Run("all - pass", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		ReportForHuman(buf, AnalyzeResult{Threshold: thr, TotalStats: coverage.Stats{}})
		assertHumanReport(t, buf.String(), 3, 0)
		assertNoUncoveredLinesInfo(t, buf.String())
	})

	t.Run("total coverage - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		ReportForHuman(buf, AnalyzeResult{Threshold: thr, TotalStats: coverage.Stats{Total: 1}})
		assertHumanReport(t, buf.String(), 2, 1)
		assertNoUncoveredLinesInfo(t, buf.String())
	})

	t.Run("file coverage - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Threshold: Threshold{File: 10}}
		statsWithError := randStats(prefix, 0, 9)
		statsNoError := randStats(prefix, 10, 100)
		allStats := mergeStats(statsWithError, statsNoError)
		result := Analyze(cfg, allStats, nil)
		ReportForHuman(buf, result)
		headReport, uncoveredReport := splitReport(t, buf.String())
		assertHumanReport(t, headReport, 0, 1)
		assertContainStats(t, headReport, statsWithError)
		assertNotContainStats(t, headReport, statsNoError)
		assertHasUncoveredLinesInfo(t, uncoveredReport,
			coverage.StatsPluckName(coverage.StatsFilterWithUncoveredLines(allStats)),
		)
		assertHasUncoveredLinesInfoWithout(t, uncoveredReport,
			coverage.StatsPluckName(coverage.StatsFilterWithCoveredLines(allStats)),
		)
	})

	t.Run("package coverage - fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Threshold: Threshold{Package: 10}}
		statsWithError := randStats(prefix, 0, 9)
		statsNoError := randStats(prefix, 10, 100)
		allStats := mergeStats(statsWithError, statsNoError)
		result := Analyze(cfg, allStats, nil)
		ReportForHuman(buf, result)
		headReport, uncoveredReport := splitReport(t, buf.String())
		assertHumanReport(t, headReport, 0, 1)
		assertContainStats(t, headReport, MakePackageStats(statsWithError))
		assertNotContainStats(t, headReport, MakePackageStats(statsNoError))
		assertNotContainStats(t, headReport, statsWithError)
		assertNotContainStats(t, headReport, statsNoError)
		assertHasUncoveredLinesInfo(t, uncoveredReport,
			coverage.StatsPluckName(coverage.StatsFilterWithUncoveredLines(allStats)),
		)
		assertHasUncoveredLinesInfoWithout(t, uncoveredReport,
			coverage.StatsPluckName(coverage.StatsFilterWithCoveredLines(allStats)),
		)
	})

	t.Run("diff - no change", func(t *testing.T) {
		t.Parallel()

		stats := randStats(prefix, 10, 100)

		buf := &bytes.Buffer{}
		cfg := Config{}
		result := Analyze(cfg, stats, stats)
		ReportForHuman(buf, result)

		assert.Contains(t, buf.String(), "Current tests coverage has not changed")
	})

	t.Run("diff - has change", func(t *testing.T) {
		t.Parallel()

		stats := randStats(prefix, 10, 100)
		base := mergeStats(make([]coverage.Stats, 0), stats)

		stats = append(stats, coverage.Stats{Name: "foo", Total: 9, Covered: 8})
		stats = append(stats, coverage.Stats{Name: "foo-new", Total: 9, Covered: 8})

		base = append(base, coverage.Stats{Name: "foo", Total: 10, Covered: 10})

		buf := &bytes.Buffer{}
		cfg := Config{}
		result := Analyze(cfg, stats, base)
		ReportForHuman(buf, result)

		assert.Contains(t, buf.String(),
			"Current tests coverage has changed with 2 lines missing coverage",
		)
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
		result := Analyze(cfg, statsNoError, nil)
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
		result := Analyze(cfg, mergeStats(statsWithError, statsNoError), nil)
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
		result := Analyze(cfg, statsNoError, nil)
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
		result := Analyze(cfg, mergeStats(statsWithError, statsNoError), nil)
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
		result := Analyze(cfg, statsNoError, nil)
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
		result := Analyze(cfg, mergeStats(statsWithError, statsNoError), nil)
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
		result := Analyze(cfg, mergeStats(statsWithError, statsNoError), nil)
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

	t.Run("writing value to output with error", func(t *testing.T) {
		err := SetOutputValue(errWriter{}, "key", "val")
		assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
		assert.Contains(t, err.Error(), "key")
	})

	t.Run("no env file", func(t *testing.T) {
		t.Setenv(GaOutputFileEnv, "")

		err := SetGithubActionOutput(AnalyzeResult{}, "")
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		testFile := t.TempDir() + "/ga.output"

		t.Setenv(GaOutputFileEnv, testFile)

		err := SetGithubActionOutput(AnalyzeResult{}, "")
		assert.NoError(t, err)

		contentBytes, err := os.ReadFile(testFile)
		assert.NoError(t, err)

		content := string(contentBytes)
		assert.Equal(t, 1, strings.Count(content, GaOutputTotalCoverage))
		assert.Equal(t, 1, strings.Count(content, GaOutputBadgeColor))
		assert.Equal(t, 1, strings.Count(content, GaOutputBadgeText))
		assert.Equal(t, 1, strings.Count(content, GaOutputReport))
	})
}

func Test_ReportUncoveredLines(t *testing.T) {
	t.Parallel()

	// when result does not pass, there should be output
	buf := &bytes.Buffer{}
	ReportUncoveredLines(buf, AnalyzeResult{
		TotalStats: coverage.Stats{Total: 1, Covered: 0},
		Threshold:  Threshold{Total: 100},
		FilesWithUncoveredLines: []coverage.Stats{
			{Name: "a.go", UncoveredLines: []int{1, 2, 3}},
			{Name: "b.go", UncoveredLines: []int{3, 5, 7}},
		},
	})
	assertHasUncoveredLinesInfo(t, buf.String(), []string{
		"a.go\t\t1-3\n",
		"b.go\t\t3 5 7\n",
	})

	// when result passes, there should be no output
	buf.Reset()
	ReportUncoveredLines(buf, AnalyzeResult{
		TotalStats: coverage.Stats{Total: 1, Covered: 1},
		Threshold:  Threshold{Total: 100},
		FilesWithUncoveredLines: []coverage.Stats{
			{Name: "a.go", UncoveredLines: []int{1, 2, 3}},
			{Name: "b.go", UncoveredLines: []int{3, 5, 7}},
		},
	})
	assert.Empty(t, buf.String())
}

func TestCompressUncoveredLines(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	CompressUncoveredLines(buf, []int{
		27, 28, 29, 30, 31, 32,
		59,
		62, 63, 64, 65, 66,
		68, 69, 70, 71, 72,
		75, 76, 77, 78, 79,
		81, 82, 83, 84, 85,
		87, 88,
	})
	assert.Equal(t, "27-32 59 62-66 68-72 75-79 81-85 87-88", buf.String())

	buf = &bytes.Buffer{}
	CompressUncoveredLines(buf, []int{
		79,
		81, 82, 83, 84, 85,
		87,
	})
	assert.Equal(t, "79 81-85 87", buf.String())
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
