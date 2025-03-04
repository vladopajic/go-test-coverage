//go:build sample
// +build sample

package sample_pkg

import "testing"

func Test_thisFuncHasCoverage(t *testing.T) {
	funcHas100PercentCoverage()
	partialCoverage()
}
