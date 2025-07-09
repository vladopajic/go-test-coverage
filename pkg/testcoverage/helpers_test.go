package testcoverage_test

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
)

func ptr[T any](t T) *T {
	return &t
}

func mergeStats(a, b []coverage.Stats) []coverage.Stats {
	r := make([]coverage.Stats, 0, len(a)+len(b))
	r = append(r, a...)
	r = append(r, b...)

	return r
}

func copyStats(s []coverage.Stats) []coverage.Stats {
	return mergeStats(make([]coverage.Stats, 0), s)
}

func randStats(localPrefix string, minc, maxc int) []coverage.Stats {
	const count = 100

	coverageGen := makeCoverageGenFn(minc, maxc)
	result := make([]coverage.Stats, 0, count)

	for {
		pkg := randPackageName(localPrefix)

		for range rand.Int31n(10) {
			total, covered := coverageGen()
			stat := coverage.Stats{
				Name:    randFileName(pkg),
				Covered: covered,
				Total:   total,
				// should have at least 1 uncovered line if has file has uncovered lines
				UncoveredLines: make([]int, min(1, total-covered)),
			}
			result = append(result, stat)

			if len(result) == count {
				return result
			}
		}
	}
}

func makeCoverageGenFn(minc, maxc int) func() (total, covered int64) {
	return func() (int64, int64) {
		tc := rand.Intn(maxc-minc+1) + minc
		if tc == 0 {
			return 0, 0
		}

		for {
			covered := int64(rand.Intn(200))
			total := int64(float64(100*covered) / float64(tc))

			cp := coverage.CoveredPercentage(total, covered)
			if cp >= minc && cp <= maxc {
				return total, covered
			}
		}
	}
}

func randPackageName(localPrefix string) string {
	if localPrefix != "" {
		localPrefix += "/"
	}

	return localPrefix + randName()
}

func randFileName(pkg string) string {
	return pkg + "/" + randName() + ".go"
}

func randName() string {
	buf := make([]byte, 10)

	_, err := crand.Read(buf)
	if err != nil {
		panic(err) //nolint:forbidigo // okay here because it is only used for tests
	}

	return hex.EncodeToString(buf)
}

func assertHumanReport(t *testing.T, content string, passCount, failCount int) {
	t.Helper()

	assert.Equal(t, passCount, strings.Count(content, "PASS"))
	assert.Equal(t, failCount, strings.Count(content, "FAIL"))
}

func assertNoFileNames(t *testing.T, content, prefix string) {
	t.Helper()

	assert.Equal(t, 0, strings.Count(content, prefix))
}

func assertContainStats(t *testing.T, content string, stats []coverage.Stats) {
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

func assertNotContainStats(t *testing.T, content string, stats []coverage.Stats) {
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

//nolint:nonamedreturns // relax
func splitReport(t *testing.T, content string) (head, uncovered string) {
	t.Helper()

	index := strings.Index(content, "Files with uncovered lines")
	if index == -1 {
		return content, ""
	}

	head = content[:index]

	content = content[index:]

	// section ends at the end of output or two \n
	index = strings.Index(content, "\n\n")
	if index == -1 {
		index = len(content)
	}

	uncovered = content[:index]

	return
}

func assertHasUncoveredLinesInfo(t *testing.T, content string, lines []string) {
	t.Helper()

	_, uncoveredReport := splitReport(t, content)
	assert.NotEmpty(t, uncoveredReport)

	for _, l := range lines {
		assert.Contains(t, uncoveredReport, l, "must contain file %v with uncovered lines", l)
	}
}

func assertHasUncoveredLinesInfoWithout(t *testing.T, content string, lines []string) {
	t.Helper()

	_, uncoveredReport := splitReport(t, content)
	assert.NotEmpty(t, uncoveredReport)

	for _, l := range lines {
		assert.NotContains(t, uncoveredReport, l, "must not contain file %v with uncovered lines", l)
	}
}

func assertNoUncoveredLinesInfo(t *testing.T, content string) {
	t.Helper()

	_, uncoveredReport := splitReport(t, content)
	assert.Empty(t, uncoveredReport)
}

func assertDiffNoChange(t *testing.T, content string) {
	t.Helper()

	assert.Contains(t, content, "No coverage changes in any files compared to the base")
}

func assertDiffChange(t *testing.T, content string, lines int) {
	t.Helper()

	//nolint:lll //relax
	str := fmt.Sprintf("Test coverage has changed in the current files, with %d lines missing coverage", lines)
	assert.Contains(t, content, str)
}

func assertDiffThreshold(t *testing.T, content string, thr float64, isSatisfied bool) {
	t.Helper()

	//nolint:lll //relax
	str := fmt.Sprintf("Coverage difference threshold (%.2f%%) satisfied:\t %s", thr, StatusStr(isSatisfied))
	assert.Contains(t, content, str)
}

func assertDiffPercentage(t *testing.T, content string, p float64) {
	t.Helper()

	str := fmt.Sprintf("Coverage difference: %.2f%%", p)
	assert.Contains(t, content, str)
}

func assertGithubActionErrorsCount(t *testing.T, content string, count int) {
	t.Helper()

	assert.Equal(t, count, strings.Count(content, "::error"))
}

func assertPrefix(t *testing.T, result AnalyzeResult, prefix string, has bool) {
	t.Helper()

	checkPrefix := func(stats []coverage.Stats) {
		for _, stat := range stats {
			assert.Equal(t, has, strings.Contains(stat.Name, prefix))
		}
	}

	checkPrefix(result.FilesBelowThreshold)
	checkPrefix(result.PackagesBelowThreshold)
}

func assertGithubOutputValues(t *testing.T, file string) {
	t.Helper()

	assertNonEmptyValue := func(t *testing.T, content, name string) {
		t.Helper()

		i := strings.Index(content, name+"")
		if i == -1 {
			t.Errorf("value [%s] not found", name)
			return
		}

		content = content[i+len(name)+1:]

		j := strings.Index(content, "\n")
		if j == -1 {
			t.Errorf("value [%s] should end with new line", name)
			return
		}

		assert.NotEmpty(t, content[:j])
	}

	contentBytes, err := os.ReadFile(file)
	assert.NoError(t, err)

	content := string(contentBytes)

	// There should be exactly 4 variables
	assert.Equal(t, 4, strings.Count(content, "="))

	// Variables should have non empty values
	assertNonEmptyValue(t, content, GaOutputTotalCoverage)
	assertNonEmptyValue(t, content, GaOutputBadgeColor)
	assertNonEmptyValue(t, content, GaOutputBadgeText)
	assertNonEmptyValue(t, content, GaOutputReport)
}

func readStats(t *testing.T, file string) []coverage.Stats {
	t.Helper()

	contentBytes, err := os.ReadFile(file)
	assert.NoError(t, err)
	assert.NotEmpty(t, contentBytes)
	stats, err := coverage.StatsDeserialize(contentBytes)
	assert.NoError(t, err)

	return stats
}
