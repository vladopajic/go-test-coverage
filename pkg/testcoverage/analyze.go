package testcoverage

import (
	"strings"
)

func Analyze(cfg Config, coverageStats []CoverageStats) AnalyzeResult {
	thr := cfg.Threshold

	filesBelowThreshold := checkCoverageStatsBelowThreshold(coverageStats, thr.File)
	packagesBelowThreshold := checkCoverageStatsBelowThreshold(
		makePackageStats(coverageStats), thr.Package,
	)
	totalStats := calcTotalStats(coverageStats)
	meetsTotalCoverage := totalStats.CoveredPercentage() >= thr.Total

	return AnalyzeResult{
		FilesBelowThreshold:    stripPrefixFromStats(filesBelowThreshold, cfg.LocalPrefix),
		PackagesBelowThreshold: stripPrefixFromStats(packagesBelowThreshold, cfg.LocalPrefix),
		MeetsTotalCoverage:     meetsTotalCoverage,
		TotalCoverage:          totalStats.CoveredPercentage(),
	}
}

func stripPrefixFromStats(coverageStats []CoverageStats, localPrefix string) []CoverageStats {
	r := make([]CoverageStats, 0, len(coverageStats))

	for _, stats := range coverageStats {
		s := CoverageStats{
			Name:    stripPrefix(stats.Name, localPrefix),
			Total:   stats.Total,
			Covered: stats.Covered,
		}
		r = append(r, s)
	}

	return r
}

func stripPrefix(name, prefix string) string {
	if prefix != "" && string(prefix[len(prefix)-1]) != "/" {
		prefix += "/"
	}

	return strings.Replace(name, prefix, "", 1)
}
