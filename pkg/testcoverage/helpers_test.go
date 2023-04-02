package testcoverage_test

import (
	crand "crypto/rand"
	"encoding/hex"
	"math"
	"math/rand"

	. "github.com/vladopajic/go-test-coverage/pkg/testcoverage"
)

func mergeCoverageStats(a, b []CoverageStats) []CoverageStats {
	r := make([]CoverageStats, 0, len(a)+len(b))
	r = append(r, a...)
	r = append(r, b...)

	return r
}

func makeCoverageStats(localPrefix string, minc, maxc int) []CoverageStats {
	const count = 100

	coverageGen := makeCoverageGenFn(minc, maxc)
	result := make([]CoverageStats, 0, count)

	for {
		pkg := randPackageName(localPrefix)

		for c := rand.Int31n(10); c >= 0; c-- { //nolint:gosec //relax
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

//nolint:gosec // relax
func makeCoverageGenFn(min, max int) func() (total, covered int64) {
	coveredPercentage := func(t, c int64) int {
		if t == 0 {
			return 0
		}

		return int(math.Round((float64(c*100) / float64(t))))
	}

	return func() (int64, int64) {
		tc := float64(rand.Intn(max-min+1) + min)

		for {
			covered := int64(rand.Intn(200))
			total := int64(math.Floor(float64(100*covered) / tc))

			cp := coveredPercentage(total, covered)
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
	buf := make([]byte, rand.Int31n(10)+10) //nolint:gosec //relax

	_, err := crand.Read(buf)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(buf)
}
