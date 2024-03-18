package testcoverage

import (
	"fmt"
	"io"
	"strings"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
)

func Check(w io.Writer, cfg Config) bool {
	stats, err := coverage.GenerateCoverageStats(coverage.Config{
		Profiles:     strings.Split(cfg.Profile, ","),
		LocalPrefix:  cfg.LocalPrefix,
		SourceDir:    cfg.SourceDir,
		ExcludePaths: cfg.Exclude.Paths,
	})
	if err != nil {
		fmt.Fprintf(w, "failed to generate coverage statistics: %v\n", err)
		return false
	}

	result := Analyze(cfg, stats)

	ReportForHuman(w, result)

	if cfg.GithubActionOutput {
		ReportForGithubAction(w, result)

		err = SetGithubActionOutput(result)
		if err != nil {
			fmt.Fprintf(w, "failed setting github action output: %v\n", err)
			return false
		}
	}

	err = generateAndSaveBadge(w, cfg, result.TotalCoverage)
	if err != nil {
		fmt.Fprintf(w, "failed to generate and save badge: %v\n", err)
		return false
	}

	return result.Pass()
}

func Analyze(cfg Config, coverageStats []coverage.Stats) AnalyzeResult {
	thr := cfg.Threshold

	overrideRules := compileOverridePathRules(cfg)

	filesBelowThreshold := checkCoverageStatsBelowThreshold(coverageStats, thr.File, overrideRules)

	packagesBelowThreshold := checkCoverageStatsBelowThreshold(
		makePackageStats(coverageStats), thr.Package, overrideRules,
	)

	totalStats := coverage.CalcTotalStats(coverageStats)
	meetsTotalCoverage := len(coverageStats) == 0 || totalStats.CoveredPercentage() >= thr.Total

	return AnalyzeResult{
		Threshold:              thr,
		FilesBelowThreshold:    filesBelowThreshold,
		PackagesBelowThreshold: packagesBelowThreshold,
		MeetsTotalCoverage:     meetsTotalCoverage,
		TotalCoverage:          totalStats.CoveredPercentage(),
	}
}
