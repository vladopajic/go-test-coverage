package testcoverage

import (
	"fmt"
	"io"
)

func Check(w io.Writer, cfg Config) (AnalyzeResult, error) {
	stats, err := GenerateCoverageStats(cfg.Profile)
	if err != nil {
		fmt.Fprintf(w, "failed to generate coverage statistics: %v\n", err)
		return AnalyzeResult{}, err
	}

	result := Analyze(cfg, stats)

	ReportForHuman(w, result, cfg)

	if cfg.GithubActionOutput {
		ReportForGithubAction(w, result, cfg)

		err := SetGithubActionOutput(result)
		if err != nil {
			fmt.Fprintf(w, "failed setting github action output: %v\n", err)
			return result, err
		}
	}

	return result, nil
}
