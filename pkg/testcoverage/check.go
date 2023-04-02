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
