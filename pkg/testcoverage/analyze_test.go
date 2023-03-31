package testcoverage_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/pkg/testcoverage"
)

func Test_Analyze(t *testing.T) {
	t.Parallel()

	localPrefix := "organization.org/" + randName()

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
			Config{LocalPrefix: localPrefix, Threshold: Threshold{Total: 10}},
			makeCoverageStats(localPrefix, 10),
		)
		assert.True(t, result.Pass())
		assertNoLocalPrefix(t, result, localPrefix)
	})

	t.Run("total coverage below threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{LocalPrefix: localPrefix, Threshold: Threshold{Total: 10}},
			makeCoverageStats(localPrefix, 9),
		)
		assert.False(t, result.Pass())
		assertNoLocalPrefix(t, result, localPrefix)
	})

	t.Run("files coverage above threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{LocalPrefix: localPrefix, Threshold: Threshold{File: 10}},
			makeCoverageStats(localPrefix, 10),
		)
		assert.True(t, result.Pass())
		assertNoLocalPrefix(t, result, localPrefix)
	})

	t.Run("files coverage below threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{LocalPrefix: localPrefix, Threshold: Threshold{File: 10}},
			mergeCoverageStats(
				makeCoverageStats(localPrefix, 9),
				makeCoverageStats(localPrefix, 10),
			),
		)
		assert.NotEmpty(t, result.FilesBelowThreshold)
		assert.Empty(t, result.PackagesBelowThreshold)
		assert.False(t, result.Pass())
		assertNoLocalPrefix(t, result, localPrefix)
	})

	t.Run("package coverage above threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{LocalPrefix: localPrefix, Threshold: Threshold{Package: 10}},
			makeCoverageStats(localPrefix, 10),
		)
		assert.True(t, result.Pass())
		assertNoLocalPrefix(t, result, localPrefix)
	})

	t.Run("package coverage below threshold", func(t *testing.T) {
		t.Parallel()

		result := Analyze(
			Config{LocalPrefix: localPrefix, Threshold: Threshold{Package: 10}},
			mergeCoverageStats(
				makeCoverageStats(localPrefix, 9),
				makeCoverageStats(localPrefix, 10),
			),
		)
		assert.Empty(t, result.FilesBelowThreshold)
		assert.NotEmpty(t, result.PackagesBelowThreshold)
		assert.False(t, result.Pass())
		assertNoLocalPrefix(t, result, localPrefix)
	})
}

func assertNoLocalPrefix(t *testing.T, result AnalyzeResult, localPrefix string) {
	t.Helper()

	noLocalPrefix := func(stats []CoverageStats) {
		for _, stat := range stats {
			assert.False(t, strings.Contains(stat.Name, localPrefix))
		}
	}

	noLocalPrefix(result.FilesBelowThreshold)
	noLocalPrefix(result.PackagesBelowThreshold)
}
