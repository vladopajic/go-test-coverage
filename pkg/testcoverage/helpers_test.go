package testcoverage_test

import (
	crand "crypto/rand"
	"encoding/hex"
	"math"
	"math/rand"

	. "github.com/vladopajic/go-test-coverage/pkg/testcoverage"
)

func mergeCoverageStats(a, b []CoverageStats) []CoverageStats {
	a = append(a, b...)
	return a
}

func makeCoverageStats(localPrefix string, coverage int) []CoverageStats {
	const minElements = 100

	result := make([]CoverageStats, 0, minElements)

	for len(result) < minElements {
		pkg := randPackageName(localPrefix)
		for c := rand.Int31n(10); c >= 0; c-- { //nolint:gosec //relax
			file := randFileName(pkg)
			result = append(result, randCoverageStats(file, coverage))
		}
	}

	return result
}

func randCoverageStats(name string, coverage int) CoverageStats {
	total := rand.Intn(500) + 1 //nolint:gosec //relax
	factor := float64(coverage) / 100

	return CoverageStats{
		Name:    name,
		Covered: int64(math.Ceil((float64(total) * factor))),
		Total:   int64(total),
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
	buf := make([]byte, rand.Int31n(10)+4) //nolint:gosec //relax

	_, err := crand.Read(buf)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(buf)
}
