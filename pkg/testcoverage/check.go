package testcoverage

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
)

func Check(w io.Writer, cfg Config) bool {
	stats, err := coverage.GenerateCoverageStats(coverage.Config{
		Profiles:     strings.Split(cfg.Profile, ","),
		LocalPrefix:  cfg.LocalPrefix,
		ExcludePaths: cfg.Exclude.Paths,
	})
	if err != nil {
		fmt.Fprintf(w, "failed to generate coverage statistics: %v\n", err)
		return false
	}

	result := Analyze(cfg, stats)

	report := reportForHuman(w, result)

	if cfg.GithubActionOutput {
		ReportForGithubAction(w, result)

		err = SetGithubActionOutput(result, report)
		if err != nil {
			fmt.Fprintf(w, "failed setting github action output: %v\n", err)
			return false
		}
	}

	err = generateAndSaveBadge(w, cfg, result.TotalStats.CoveredPercentage())
	if err != nil {
		fmt.Fprintf(w, "failed to generate and save badge: %v\n", err)
		return false
	}

	return result.Pass()
}

func reportForHuman(w io.Writer, result AnalyzeResult) string {
	buffer := &bytes.Buffer{}
	out := bufio.NewWriter(buffer)

	ReportForHuman(out, result)
	out.Flush()

	w.Write(buffer.Bytes()) //nolint:errcheck // relax

	return buffer.String()
}

func Analyze(cfg Config, coverageStats []coverage.Stats) AnalyzeResult {
	thr := cfg.Threshold
	overrideRules := compileOverridePathRules(cfg)

	return AnalyzeResult{
		Threshold:           thr,
		FilesBelowThreshold: checkCoverageStatsBelowThreshold(coverageStats, thr.File, overrideRules),
		PackagesBelowThreshold: checkCoverageStatsBelowThreshold(
			makePackageStats(coverageStats), thr.Package, overrideRules,
		),
		TotalStats: coverage.CalcTotalStats(coverageStats),
	}
}
