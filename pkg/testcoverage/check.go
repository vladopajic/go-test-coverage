package testcoverage

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/logger"
)

//nolint:maintidx, nonamedreturns // relax
func Check(wout io.Writer, cfg Config) (passed bool, haderr bool) {
	buffer := &bytes.Buffer{}
	w := bufio.NewWriter(buffer)
	//nolint:errcheck // relax
	defer func() {
		if cfg.Debug {
			wout.Write(logger.Bytes())
			wout.Write([]byte("-------------------------\n\n"))
		}

		w.Flush()
		wout.Write(buffer.Bytes())
	}()

	logger.L.Debug().Msg("running check...")
	logger.L.Debug().Any("config", cfg).Msg("using configuration")

	currentStats, err := GenerateCoverageStats(cfg)
	if err != nil {
		logger.L.Error().Err(err).Msg("failed to generate coverage statistics")
		return false, true
	}

	err = saveCoverageBreakdown(cfg, currentStats)
	if err != nil {
		logger.L.Error().Err(err).Msg("failed to save coverage breakdown")
		return false, true
	}

	baseStats, err := loadBaseCoverageBreakdown(cfg)
	if err != nil {
		logger.L.Error().Err(err).Msg("failed to load base coverage breakdown")
		return false, true
	}

	result := Analyze(cfg, currentStats, baseStats)

	report := reportForHuman(w, result)

	if cfg.GithubActionOutput {
		ReportForGithubAction(w, result)

		err = SetGithubActionOutput(result, report)
		if err != nil {
			logger.L.Error().Err(err).Msg("failed setting github action output")
			return false, true
		}

		if cfg.LocalPrefixDeprecated != "" { // coverage-ignore
			reportGHWarning(w, "Deprecated option",
				"local-prefix option is deprecated since v2.13.0, you can safely remove setting this option")
		}
	}

	err = generateAndSaveBadge(w, cfg, result.TotalStats.CoveredPercentage())
	if err != nil {
		logger.L.Error().Err(err).Msg("failed to generate and save badge")
		return false, true
	}

	return result.Pass(), false
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
		ExcludePaths: cfg.Exclude.Paths,
		SourceDir:    cfg.SourceDir,
	})
}

func Analyze(cfg Config, current, base []coverage.Stats) AnalyzeResult {
	thr := cfg.Threshold
	overrideRules := compileOverridePathRules(cfg)
	hasFileOverrides, hasPackageOverrides := detectOverrides(cfg.Override)

	return AnalyzeResult{
		Threshold:           thr,
		HasFileOverrides:    hasFileOverrides,
		HasPackageOverrides: hasPackageOverrides,
		FilesBelowThreshold: checkCoverageStatsBelowThreshold(current, thr.File, overrideRules),
		PackagesBelowThreshold: checkCoverageStatsBelowThreshold(
			makePackageStats(current), thr.Package, overrideRules,
		),
		FilesWithUncoveredLines: coverage.StatsFilterWithUncoveredLines(current),
		TotalStats:              coverage.StatsCalcTotal(current),
		HasBaseBreakdown:        len(base) > 0,
		Diff:                    calculateStatsDiff(current, base),
	}
}

func detectOverrides(overrides []Override) (bool, bool) {
	hasFileOverrides := false
	hasPackageOverrides := false

	for _, override := range overrides {
		if strings.HasSuffix(override.Path, ".go") || strings.HasSuffix(override.Path, ".go$") {
			hasFileOverrides = true
		} else {
			hasPackageOverrides = true
		}
	}

	return hasFileOverrides, hasPackageOverrides
}

func saveCoverageBreakdown(cfg Config, stats []coverage.Stats) error {
	if cfg.BreakdownFileName == "" {
		return nil
	}

	//nolint:mnd,wrapcheck,gosec // relax
	return os.WriteFile(cfg.BreakdownFileName, coverage.StatsSerialize(stats), 0o644)
}

func loadBaseCoverageBreakdown(cfg Config) ([]coverage.Stats, error) {
	if cfg.Diff.BaseBreakdownFileName == "" {
		return nil, nil
	}

	data, err := os.ReadFile(cfg.Diff.BaseBreakdownFileName)
	if err != nil {
		return nil, fmt.Errorf("reading file content failed: %w", err)
	}

	stats, err := coverage.StatsDeserialize(data)
	if err != nil {
		return nil, fmt.Errorf("parsing file failed: %w", err)
	}

	return stats, nil
}
