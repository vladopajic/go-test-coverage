package testcoverage

import (
	"math"
	"strings"
)

type AnalyzeResult struct {
	FilesBelowThreshold    []CoverageStats
	PackagesBelowThreshold []CoverageStats
	MeetsTotalCoverage     bool
	TotalCoverage          int
}

func (r *AnalyzeResult) Pass() bool {
	return r.MeetsTotalCoverage &&
		len(r.FilesBelowThreshold) == 0 &&
		len(r.PackagesBelowThreshold) == 0
}

type CoverageStats struct {
	Name    string
	Total   int64
	Covered int64
}

func (s *CoverageStats) CoveredPercentage() int {
	if s.Total == 0 {
		return 0
	}

	//nolint:gomnd // relax
	return int(math.Round((float64(s.Covered*100) / float64(s.Total))))
}

func checkCoverageStatsBelowThreshold(
	coverageStats []CoverageStats,
	threshold int,
) []CoverageStats {
	belowThreshold := make([]CoverageStats, 0)

	for _, stats := range coverageStats {
		if stats.CoveredPercentage() < threshold {
			belowThreshold = append(belowThreshold, stats)
		}
	}

	return belowThreshold
}

func calcTotalStats(coverageStats []CoverageStats) CoverageStats {
	totalStats := CoverageStats{}

	for _, stats := range coverageStats {
		totalStats.Total += stats.Total
		totalStats.Covered += stats.Covered
	}

	return totalStats
}

func makePackageStats(coverageStats []CoverageStats) []CoverageStats {
	packageStats := make(map[string]CoverageStats)

	for _, stats := range coverageStats {
		pkg := packageForFile(stats.Name)

		var pkgStats CoverageStats
		if s, ok := packageStats[pkg]; ok {
			pkgStats = s
		} else {
			pkgStats = CoverageStats{Name: pkg}
		}

		pkgStats.Total += stats.Total
		pkgStats.Covered += stats.Covered
		packageStats[pkg] = pkgStats
	}

	packageStatsSlice := make([]CoverageStats, 0, len(packageStats))
	for _, stats := range packageStats {
		packageStatsSlice = append(packageStatsSlice, stats)
	}

	return packageStatsSlice
}

func packageForFile(filename string) string {
	i := strings.LastIndex(filename, "/")
	if i == -1 {
		return filename
	}

	return filename[:i]
}
