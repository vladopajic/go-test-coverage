package testcoverage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

		if cfg.LocalPrefixDeprecated != "" { // coverage-ignore
			reportGHWarning(w, "Deprecated option",
				"local-prefix option is deprecated since v2.13.0, you can safely remove setting this option")
		}
	}

	err = generateAndSaveBadge(w, cfg, result.TotalStats.CoveredPercentage())
	if err != nil {
		fmt.Fprintf(w, "failed to generate and save badge: %v\n", err)
		return false
	}

	// New: post coverage comment to PR if enabled.
	if cfg.PostCoverageComment {
		if err := postPRComment(w, cfg, result, report); err != nil {
			fmt.Fprintf(w, "failed to post PR comment: %v\n", err)
			// do not fail the check due to comment-posting issues
		}
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

// New helper function to post a comment to the PR using GitHub API.
func postPRComment(w io.Writer, cfg Config, result AnalyzeResult, report string) error {
	// Expecting a GitHub token in the environment.
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN not set")
	}

	// Read the event payload to extract the PR number.
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		// Not running in GitHub Actions or no event payload available.
		return nil
	}
	data, err := os.ReadFile(eventPath)
	if err != nil {
		return fmt.Errorf("failed reading GITHUB_EVENT_PATH: %w", err)
	}
	// Define a minimal structure to get the pull_request number.
	var event struct {
		PullRequest struct {
			Number int `json:"number"`
		} `json:"pull_request"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed parsing event payload: %w", err)
	}
	if event.PullRequest.Number == 0 {
		// Not a pull request event.
		return nil
	}

	// Get repository information from environment.
	repoStr := os.Getenv("GITHUB_REPOSITORY") // format: owner/repo
	if repoStr == "" {
		return fmt.Errorf("GITHUB_REPOSITORY not set")
	}
	parts := strings.Split(repoStr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("GITHUB_REPOSITORY formatted invalidly")
	}
	owner, repoName := parts[0], parts[1]
	commentURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repoName, event.PullRequest.Number)

	// Construct the comment body.
	commentBody := fmt.Sprintf("Coverage Report:\n```\n%s\n```", report)
	payload, err := json.Marshal(map[string]string{
		"body": commentBody,
	})
	if err != nil {
		return fmt.Errorf("failed marshaling payload: %w", err)
	}

	req, err := http.NewRequest("POST", commentURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed creating request: %w", err)
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to post comment, status: %d, response: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
