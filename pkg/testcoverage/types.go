package testcoverage

import (
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
	return CoveredPercentage(s.Total, s.Covered)
}

//nolint:gomnd // relax
func CoveredPercentage(total, covered int64) int {
	if total == 0 {
		return 0
	}

	if covered == total {
		return 100
	}

	return int(float64(covered*100) / float64(total))
}

func checkCoverageStatsBelowThreshold(
	coverageStats []CoverageStats,
	threshold int,
) []CoverageStats {
	var belowThreshold []CoverageStats

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

func stripPrefix(name, prefix string) string {
	if prefix == "" {
		return name
	}

	if string(prefix[len(prefix)-1]) != "/" {
		prefix += "/"
	}

	return strings.Replace(name, prefix, "", 1)
}
