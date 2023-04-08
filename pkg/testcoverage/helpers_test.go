package testcoverage_test

import (
	crand "crypto/rand"
	"encoding/hex"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
)

func mergeStats(a, b []CoverageStats) []CoverageStats {
	r := make([]CoverageStats, 0, len(a)+len(b))
	r = append(r, a...)
	r = append(r, b...)

	return r
}

func randStats(localPrefix string, minc, maxc int) []CoverageStats {
	const count = 100

	coverageGen := makeCoverageGenFn(minc, maxc)
	result := make([]CoverageStats, 0, count)

	for {
		pkg := randPackageName(localPrefix)

		for c := rand.Int31n(10); c >= 0; c-- {
			total, covered := coverageGen()
			stat := CoverageStats{
				Name:    randFileName(pkg),
				Covered: covered,
				Total:   total,
			}
			result = append(result, stat)

			if len(result) == count {
				return result
			}
		}
	}
}

func makeCoverageGenFn(min, max int) func() (total, covered int64) {
	return func() (int64, int64) {
		tc := rand.Intn(max-min+1) + min
		if tc == 0 {
			return 0, 0
		}

		for {
			covered := int64(rand.Intn(200))
			total := int64(float64(100*covered) / float64(tc))

			cp := CoveredPercentage(total, covered)
			if cp >= min && cp <= max {
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
		panic(err)
	}

	return hex.EncodeToString(buf)
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

func assertPrefix(t *testing.T, result AnalyzeResult, prefix string, has bool) {
	t.Helper()

	checkPrefix := func(stats []CoverageStats) {
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

	// There should be exactly 3 variables
	assert.Equal(t, 3, strings.Count(content, "="))

	// Variables should have non empty values
	assertNonEmptyValue(t, content, GaOutputTotalCoverage)
	assertNonEmptyValue(t, content, GaOutputBadgeColor)
	assertNonEmptyValue(t, content, GaOutputBadgeText)
}
