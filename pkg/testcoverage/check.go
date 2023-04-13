package testcoverage

import (
	"fmt"
	"io"
)

func Check(w io.Writer, cfg Config) (AnalyzeResult, error) {
	stats, err := GenerateCoverageStats(cfg)
	if err != nil {
		fmt.Fprintf(w, "failed to generate coverage statistics: %v\n", err)
		return AnalyzeResult{}, err
	}

	result := Analyze(cfg, stats)

	ReportForHuman(w, result, cfg.Threshold)

	if cfg.GithubActionOutput {
		ReportForGithubAction(w, result, cfg.Threshold)

		err := SetGithubActionOutput(result)
		if err != nil {
			fmt.Fprintf(w, "failed setting github action output: %v\n", err)
			return result, err
		}
	}

	return result, nil
}

func Analyze(cfg Config, coverageStats []CoverageStats) AnalyzeResult {
	thr := cfg.Threshold

	filesBelowThreshold := checkCoverageStatsBelowThreshold(coverageStats, thr.File)
	packagesBelowThreshold := checkCoverageStatsBelowThreshold(
		makePackageStats(coverageStats), thr.Package,
	)
	totalStats := calcTotalStats(coverageStats)
	meetsTotalCoverage := totalStats.CoveredPercentage() >= thr.Total

	return AnalyzeResult{
		FilesBelowThreshold:    filesBelowThreshold,
		PackagesBelowThreshold: packagesBelowThreshold,
		MeetsTotalCoverage:     meetsTotalCoverage,
		TotalCoverage:          totalStats.CoveredPercentage(),
	}
}
