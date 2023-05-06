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

	ReportForHuman(w, result)

	if cfg.GithubActionOutput {
		ReportForGithubAction(w, result)

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

	overrideRules := compileOverridePathRules(cfg)

	filesBelowThreshold := checkCoverageStatsBelowThreshold(coverageStats, thr.File, overrideRules)

	packagesBelowThreshold := checkCoverageStatsBelowThreshold(
		makePackageStats(coverageStats), thr.Package, overrideRules,
	)

	totalStats := calcTotalStats(coverageStats)
	meetsTotalCoverage := len(coverageStats) == 0 || totalStats.CoveredPercentage() >= thr.Total

	return AnalyzeResult{
		Threshold:              thr,
		FilesBelowThreshold:    filesBelowThreshold,
		PackagesBelowThreshold: packagesBelowThreshold,
		MeetsTotalCoverage:     meetsTotalCoverage,
		TotalCoverage:          totalStats.CoveredPercentage(),
	}
}
