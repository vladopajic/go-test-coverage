package testcoverage_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/pkg/testcoverage"
)

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
