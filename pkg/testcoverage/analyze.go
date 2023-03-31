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

func Analyze(cfg Config, coverageStats []CoverageStats) AnalyzeResult {
	thr := cfg.Threshold

	filesBelowThreshold := checkCoverageStatsBelowThreshold(coverageStats, thr.File)
	packagesBelowThreshold := checkCoverageStatsBelowThreshold(
		makePackageStats(coverageStats), thr.Package,
	)
	totalStats := calcTotalStats(coverageStats)
	meetsTotalCoverage := totalStats.CoveredPercentage() >= thr.Total

	localPrefix := cfg.LocalPrefix
	if localPrefix != "" && (strings.LastIndex(localPrefix, "/") != len(localPrefix)-1) {
		localPrefix += "/"
	}

	return AnalyzeResult{
		FilesBelowThreshold:    stripLocalPrefix(filesBelowThreshold, localPrefix),
		PackagesBelowThreshold: stripLocalPrefix(packagesBelowThreshold, localPrefix),
		MeetsTotalCoverage:     meetsTotalCoverage,
		TotalCoverage:          totalStats.CoveredPercentage(),
	}
}

func stripLocalPrefix(coverageStats []CoverageStats, localPrefix string) []CoverageStats {
	for i, stats := range coverageStats {
		coverageStats[i].Name = strings.Replace(stats.Name, localPrefix, "", 1)
	}

	return coverageStats
}
