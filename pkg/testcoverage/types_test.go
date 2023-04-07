package testcoverage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
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
