package coverage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
)

func TestCoveredPercentage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		percentage int
		total      int64
		covered    int64
	}{
		{percentage: 0, total: 0, covered: 0},
		{percentage: 0, total: 0, covered: 1},
		{percentage: 100, total: 1, covered: 1},
		{percentage: 10, total: 10, covered: 1},
		{percentage: 22, total: 9, covered: 2}, // 22.222.. should round down to 22
		{percentage: 66, total: 9, covered: 6}, // 66.666.. should round down to 66
	}

	for _, tc := range tests {
		assert.Equal(t, tc.percentage, CoveredPercentage(tc.total, tc.covered))
	}
}

func TestStatStr(t *testing.T) {
	t.Parallel()

	assert.Equal(t, " 0.0% (0/0)", Stats{}.Str())
	assert.Equal(t, " 9.1% (1/11)", Stats{Covered: 1, Total: 11}.Str())
	assert.Equal(t, "22.2% (2/9)", Stats{Covered: 2, Total: 9}.Str())
	assert.Equal(t, "100% (10/10)", Stats{Covered: 10, Total: 10}.Str())
}
