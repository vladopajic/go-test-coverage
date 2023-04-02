package testcoverage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"text/tabwriter"
)

func ReportForHuman(w io.Writer, result AnalyzeResult, cfg Config) {
	thr := cfg.Threshold

	out := bufio.NewWriter(w)
	defer out.Flush()

	{
		fmt.Fprintf(out, "Files meeting coverage threshold of (%d%%):\t", thr.File)
		if len(result.FilesBelowThreshold) > 0 {
			fmt.Fprintf(out, "FAIL")
			reportIssuesForHuman(out, result.FilesBelowThreshold)
		} else {
			fmt.Fprintf(out, "PASS")
		}
	}

	{
		fmt.Fprintf(out, "\nPackages meeting coverage threshold of (%d%%):\t", thr.Package)
		if len(result.PackagesBelowThreshold) > 0 {
			fmt.Fprintf(out, "FAIL")
			reportIssuesForHuman(out, result.PackagesBelowThreshold)
		} else {
			fmt.Fprintf(out, "PASS")
		}
	}

	{
		fmt.Fprintf(out, "\nTotal coverage meeting the threshold of (%d%%):\t", thr.Total)
		if !result.MeetsTotalCoverage {
			fmt.Fprintf(out, "FAIL")
		} else {
			fmt.Fprintf(out, "PASS")
		}
	}

	fmt.Fprintf(out, "\nTotal test coverage: %d%%\n", result.TotalCoverage)
}

func reportIssuesForHuman(w io.Writer, coverageStats []CoverageStats) {
	tabber := tabwriter.NewWriter(w, 1, 8, 1, '\t', 0) //nolint:gomnd // relax
	defer tabber.Flush()

	fmt.Fprintf(tabber, "\n\nIssues with:")

	for _, stats := range coverageStats {
		fmt.Fprintf(tabber, "\n%s\t%d%%", stats.Name, stats.CoveredPercentage())
	}

	fmt.Fprintf(tabber, "\n")
}

func ReportForGithubAction(w io.Writer, result AnalyzeResult, cfg Config) {
	out := bufio.NewWriter(w)
	defer out.Flush()

	reportLineError := func(file, title, msg string) {
		fmt.Fprintf(out, "::error file=%s,title=%s,line=1::%s\n", file, title, msg)
	}
	reportError := func(title, msg string) {
		fmt.Fprintf(out, "::error title=%s::%s\n", title, msg)
	}

	for _, stats := range result.FilesBelowThreshold {
		title := "File test coverage below threshold"
		c := stats.CoveredPercentage()
		t := cfg.Threshold.File
		msg := fmt.Sprintf("coverage: %d%%; threshold: %d%%", c, t)
		reportLineError(stats.Name, title, msg)
	}

	for _, stats := range result.PackagesBelowThreshold {
		title := "Package test coverage below threshold"
		c := stats.CoveredPercentage()
		t := cfg.Threshold.Package
		msg := fmt.Sprintf("package: %s; coverage: %d%%; threshold: %d%%", stats.Name, c, t)
		reportError(title, msg)
	}

	if !result.MeetsTotalCoverage {
		title := "Total test coverage below threshold"
		c := result.TotalCoverage
		t := cfg.Threshold.Total
		msg := fmt.Sprintf("coverage: %d%%; threshold: %d%%", c, t)
		reportError(title, msg)
	}
}

const (
	gaOutputFileEnv       = "GITHUB_OUTPUT"
	gaOutputTotalCoverage = "total_coverage"
	gaOutputBadgeColor    = "badge_color"
	gaOutputBadgeText     = "badge_text"
)

func SetGithubActionOutput(result AnalyzeResult) error {
	file, err := openGitHubOutput(os.Getenv(gaOutputFileEnv))
	if err != nil {
		return fmt.Errorf("could not open GitHub output file: %w", err)
	}
	defer file.Close()

	totalText := strconv.Itoa(result.TotalCoverage)

	return errors.Join(
		setOutput(file, gaOutputTotalCoverage, totalText),
		setOutput(file, gaOutputBadgeColor, coverageColor(result.TotalCoverage)),
		setOutput(file, gaOutputBadgeText, totalText+"%"),
	)
}

func openGitHubOutput(p string) (io.WriteCloser, error) {
	//nolint:gomnd,wrapcheck //relax
	return os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
}

func setOutput(w io.Writer, name, value string) error {
	if _, err := w.Write([]byte(fmt.Sprintf("%s=%s\n", name, value))); err != nil {
		return fmt.Errorf("failed setting github output [var=%s]: %w", name, err)
	}

	return nil
}

func coverageColor(coverage int) string {
	if coverage >= 100 {
		return "#44cc11" // green
	} else if coverage >= 90 {
		return "#97ca00" // light green
	} else if coverage >= 80 {
		return "#dfb317" // yellow
	} else if coverage >= 70 {
		return "#fa7739" // orange
	}
	return "#e05d44" // red
}
