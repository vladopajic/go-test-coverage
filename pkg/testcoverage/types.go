package testcoverage

import (
	"strings"

	"golang.org/x/exp/maps"

	"github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage/coverage"
)

type AnalyzeResult struct {
	Threshold              Threshold
	FilesBelowThreshold    []coverage.Stats
	PackagesBelowThreshold []coverage.Stats
	TotalStats             coverage.Stats
	HasBaseBreakdown       bool
	Diff                   []FileCoverageDiff
}

func (r *AnalyzeResult) Pass() bool {
	return r.MeetsTotalCoverage() &&
		len(r.FilesBelowThreshold) == 0 &&
		len(r.PackagesBelowThreshold) == 0
}

func (r *AnalyzeResult) MeetsTotalCoverage() bool {
	return r.TotalStats.Total == 0 || r.TotalStats.CoveredPercentage() >= r.Threshold.Total
}

func packageForFile(filename string) string {
	i := strings.LastIndex(filename, "/")
	if i == -1 {
		return filename
	}

	return filename[:i]
}

func checkCoverageStatsBelowThreshold(
	coverageStats []coverage.Stats,
	threshold int,
	overrideRules []regRule,
) []coverage.Stats {
	var belowThreshold []coverage.Stats

	for _, s := range coverageStats {
		thr := threshold
		if override, ok := matches(overrideRules, s.Name); ok {
			thr = override
		}

		if s.CoveredPercentage() < thr {
			s.Threshold = thr
			belowThreshold = append(belowThreshold, s)
		}
	}

	return belowThreshold
}

func makePackageStats(coverageStats []coverage.Stats) []coverage.Stats {
	packageStats := make(map[string]coverage.Stats)

	for _, stats := range coverageStats {
		pkg := packageForFile(stats.Name)

		var pkgStats coverage.Stats
		if s, ok := packageStats[pkg]; ok {
			pkgStats = s
		} else {
			pkgStats = coverage.Stats{Name: pkg}
		}

		pkgStats.Total += stats.Total
		pkgStats.Covered += stats.Covered
		packageStats[pkg] = pkgStats
	}

	return maps.Values(packageStats)
}

type FileCoverageDiff struct {
	Current coverage.Stats
	Base    *coverage.Stats
}

func calculateStatsDiff(current, base []coverage.Stats) []FileCoverageDiff {
	res := make([]FileCoverageDiff, 0)
	baseSearchMap := coverage.StatsSearchMap(base)

	for _, s := range current {
		sul := s.UncoveredLines()
		if sul == 0 {
			continue
		}

		if b, found := baseSearchMap[s.Name]; found {
			if sul != b.UncoveredLines() {
				res = append(res, FileCoverageDiff{Current: s, Base: &b})
			}
		} else {
			res = append(res, FileCoverageDiff{Current: s})
		}
	}

	return res
}

func TotalLinesDiff(diff []FileCoverageDiff) int {
	r := 0
	for _, d := range diff {
		r += d.Current.UncoveredLines()
	}

	return r
}
