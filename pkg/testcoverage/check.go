package testcoverage

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
)

func Check(w io.Writer, cfg Config) bool {
	currentStats, err := GenerateCoverageStats(cfg)
	if err != nil {
		fmt.Fprintf(w, "failed to generate coverage statistics: %v\n", err)
		return false
	}

	err = saveCoverageBreakdown(cfg, currentStats)
	if err != nil {
		fmt.Fprintf(w, "failed to save coverage breakdown: %v\n", err)
		return false
	}

	baseStats, err := loadBaseCoverageBreakdown(cfg)
	if err != nil {
		fmt.Fprintf(w, "failed to load base coverage breakdown: %v\n", err)
		return false
	}

	result := Analyze(cfg, currentStats, baseStats)

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

func GenerateCoverageStats(cfg Config) ([]coverage.Stats, error) {
	return coverage.GenerateCoverageStats(coverage.Config{ //nolint:wrapcheck // err wrapped above
		Profiles:     strings.Split(cfg.Profile, ","),
		LocalPrefix:  cfg.LocalPrefix,
		ExcludePaths: cfg.Exclude.Paths,
	})
}

func Analyze(cfg Config, current, base []coverage.Stats) AnalyzeResult {
	thr := cfg.Threshold
	overrideRules, hasOverrides := compileOverridePathRules(cfg)

	return AnalyzeResult{
		Threshold:           thr,
		HasOverrides:        hasOverrides,
		FilesBelowThreshold: checkCoverageStatsBelowThreshold(current, thr.File, overrideRules),
		PackagesBelowThreshold: checkCoverageStatsBelowThreshold(
			makePackageStats(current), thr.Package, overrideRules,
		),
		TotalStats:       coverage.CalcTotalStats(current),
		HasBaseBreakdown: len(base) > 0,
		Diff:             calculateStatsDiff(current, base),
	}
}

func saveCoverageBreakdown(cfg Config, stats []coverage.Stats) error {
	if cfg.BreakdownFileName == "" {
		return nil
	}

	//nolint:mnd,wrapcheck,gosec // relax
	return os.WriteFile(cfg.BreakdownFileName, coverage.SerializeStats(stats), 0o644)
}

func loadBaseCoverageBreakdown(cfg Config) ([]coverage.Stats, error) {
	if cfg.Diff.BaseBreakdownFileName == "" {
		return nil, nil
	}

	data, err := os.ReadFile(cfg.Diff.BaseBreakdownFileName)
	if err != nil {
		return nil, fmt.Errorf("reading file content failed: %w", err)
	}

	stats, err := coverage.DeserializeStats(data)
	if err != nil {
		return nil, fmt.Errorf("parsing file failed: %w", err)
	}

	return stats, nil
}
