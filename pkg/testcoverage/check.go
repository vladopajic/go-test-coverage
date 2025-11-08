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

//nolint:maintidx // relax
func Check(wout io.Writer, cfg Config) (bool, error) {
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

	handleErr := func(err error, msg string) (bool, error) {
		logger.L.Error().Err(err).Msg(msg)
		return false, fmt.Errorf("%s: %w", msg, err)
	}

	logger.L.Info().Msg("running check...")
	logger.L.Info().Any("config", cfg.Redacted()).Msg("using configuration")

	currentStats, err := GenerateCoverageStats(cfg)
	if err != nil {
		return handleErr(err, "failed to generate coverage statistics")
	}

	err = saveCoverageBreakdown(cfg, currentStats)
	if err != nil {
		return handleErr(err, "failed to save coverage breakdown")
	}

	baseStats, err := loadBaseCoverageBreakdown(cfg)
	if err != nil {
		return handleErr(err, "failed to load base coverage breakdown")
	}

	result := Analyze(cfg, currentStats, baseStats)

	report := reportForHuman(w, result)

	if cfg.GithubActionOutput {
		ReportForGithubAction(w, result)

		err = SetGithubActionOutput(result, report)
		if err != nil {
			return handleErr(err, "failed setting github action output")
		}

		if cfg.LocalPrefixDeprecated != "" { // coverage-ignore
			//nolint:lll // relax
			msg := "`local-prefix` option is deprecated since v2.13.0, you can safely remove setting this option"
			logger.L.Warn().Msg(msg)
			reportGHWarning(w, "Deprecated option", msg)
		}
	}

	err = generateAndSaveBadge(w, cfg, result.TotalStats.CoveredPercentage())
	if err != nil {
		return handleErr(err, "failed to generate and save badge")
	}

	return result.Pass(), nil
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
		Profiles:               strings.Split(cfg.Profile, ","),
		ExcludePaths:           cfg.Exclude.Paths,
		SourceDir:              cfg.SourceDir,
		ForceAnnotationComment: cfg.ForceAnnotationComment,
	})
}

func Analyze(cfg Config, current, base []coverage.Stats) AnalyzeResult {
	thr := cfg.Threshold
	overrideRules := compileOverridePathRules(cfg)
	hasFileOverrides, hasPackageOverrides := detectOverrides(cfg.Override)

	var filesWithMissingExplanations []coverage.Stats
	if cfg.ForceAnnotationComment {
		filesWithMissingExplanations = coverage.StatsFilterWithMissingExplanations(current)
	}

	return AnalyzeResult{
		Threshold:           thr,
		DiffThreshold:       cfg.Diff.Threshold,
		HasFileOverrides:    hasFileOverrides,
		HasPackageOverrides: hasPackageOverrides,
		FilesBelowThreshold: checkCoverageStatsBelowThreshold(current, thr.File, overrideRules),
		PackagesBelowThreshold: checkCoverageStatsBelowThreshold(
			makePackageStats(current), thr.Package, overrideRules,
		),
		FilesWithUncoveredLines:      coverage.StatsFilterWithUncoveredLines(current),
		FilesWithMissingExplanations: filesWithMissingExplanations,
		TotalStats:                   coverage.StatsCalcTotal(current),
		HasBaseBreakdown:             len(base) > 0,
		Diff:                         calculateStatsDiff(current, base),
		DiffPercentage:               TotalPercentageDiff(current, base),
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
		return nil, fmt.Errorf("deserializing stats file failed: %w", err)
	}

	return stats, nil
}
